package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var (
	sourceEncodingArg string
	targetEncodingArg string
	sourceEncoding    encoding.Encoding
	targetEncoding    encoding.Encoding

	doReplace = false
	doVerbose = false
)
var cmd = &cobra.Command{
	Use:   "Work with files:\n    transcode [-r] [-s encoding] [-t encoding] [-v] files...\n  Work with stdin:\n    cat file | transcode",
	Short: "Translate text encoding",
	RunE: func(md *cobra.Command, args []string) (err error) {
		sourceEncoding, err = parseEncoding(sourceEncodingArg)
		if err != nil {
			return fmt.Errorf("parse source-encoding failed: %w", err)
		}
		targetEncoding, err = parseEncoding(targetEncodingArg)
		if err != nil {
			return fmt.Errorf("parse target-encoding failed: %w", err)
		}

		if len(args) == 0 {
			args = []string{"-"}
			doVerbose = false
		}

		if doVerbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		for _, file := range args {
			logrus.Debugf("process %s start", file)
			err = processFile(file)
			if err != nil {
				logrus.WithError(err).Errorf("process %s error", file)
			}
			logrus.Debugf("process %s end", file)
		}

		return
	},
}

func init() {
	cmd.Flags().SortFlags = false
	cmd.PersistentFlags().SortFlags = false
	cmd.PersistentFlags().BoolVarP(
		&doReplace,
		"replace-source",
		"r",
		false,
		"replace/overwrite source file",
	)
	cmd.PersistentFlags().StringVarP(
		&sourceEncodingArg,
		"source-encoding",
		"s",
		"GBK",
		"source encoding",
	)
	cmd.PersistentFlags().StringVarP(
		&targetEncodingArg,
		"target-encoding",
		"t",
		"UTF8",
		"target encoding",
	)
	cmd.PersistentFlags().BoolVarP(
		&doVerbose,
		"verbose",
		"v",
		false,
		"verbose",
	)

	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		//NoColors:        true,
		HideKeys:    true,
		CallerFirst: true,
	})
}
func main() {
	_ = cmd.Execute()
}
func parseEncoding(encoding string) (enc encoding.Encoding, err error) {
	enc, err = htmlindex.Get(encoding)
	if err != nil {
		err = fmt.Errorf("invalid encoding: %s", encoding)
	}
	return
}
func processFile(file string) (err error) {
	source, target, isDiskOp := os.Stdin, os.Stdout, file != "-"

	if isDiskOp {
		var abs string
		{
			logger := logrus.WithField("step", "RESOLVE")
			logger.Debugf("resolve %s", file)
			abs, err = filepath.Abs(file)
			if err != nil {
				err = fmt.Errorf("cannot parse file path %s", file)
				return
			}
			logger.Debugf("resolve %s done: %s", file, abs)
		}
		{
			logger := logrus.WithField("step", "OPEN")
			logger.Debugf("open %s", abs)
			source, err = os.Open(abs)
			if err == nil {
				defer source.Close()
			} else {
				return
			}
			logger.Debugf("open %s ok", abs)
		}
		{
			logger := logrus.WithField("step", "OPEN")
			logger.Debugf("create temp file")
			target, err = ioutil.TempFile("", fmt.Sprintf("file_%s.", filepath.Base(source.Name())))
			if err == nil {
				defer target.Close()
			} else {
				return
			}
			logger.Debugf("create temp file %s ok", target.Name())
		}
	}

	logger := logrus.WithField("step", "TRANSFER")
	logger.Debugf("transfer %s => %s", source.Name(), target.Name())
	_, err = io.Copy(
		transform.NewWriter(target, targetEncoding.NewEncoder()),
		transform.NewReader(source, sourceEncoding.NewDecoder()),
	)
	if err != nil {
		err = fmt.Errorf("translate %s => %s failed: %s", source.Name(), target.Name(), err)
		return
	}
	logger.Debugf("transfer %s => %s done", source.Name(), target.Name())

	if isDiskOp {
		targetRename := source.Name() + ".out"

		if doReplace {
			logger := logrus.WithField("step", "REMOVE")
			logger.Debugf("remove %s", source.Name())
			err = os.Remove(source.Name())
			if err != nil {
				err = fmt.Errorf("remove %s failed: %w", source.Name(), err)
				return
			}
			logger.Debugf("remove %s done", source.Name())

			targetRename = source.Name()
		}

		logger := logrus.WithField("step", "RENAME")
		{
			logger.Debugf("rename %s => %s", target.Name(), targetRename)
			err = os.Rename(target.Name(), targetRename)
			if err != nil {
				err = fmt.Errorf("rename %s => %s failed: %w", target.Name(), targetRename, err)
				return
			}
			logger.Debugf("rename %s => %s done", target.Name(), targetRename)
		}

		logrus.Infof("save into %s", targetRename)
	}

	return
}
