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

	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var (
	flagSourceEncoding string
	flagTargetEncoding string

	flagReplace bool
	flagVerbose bool

	sourceEncoding encoding.Encoding
	targetEncoding encoding.Encoding

	cmd = &cobra.Command{
		Use:   "Work with files:\n    transcode [-r] [-s encoding] [-t encoding] [-v] files...\n  Work with stdin:\n    cat file | transcode",
		Short: "Translate text encoding",
		RunE: func(md *cobra.Command, args []string) (err error) {
			sourceEncoding, err = parseEncoding(flagSourceEncoding)
			if err != nil {
				return fmt.Errorf("parse source-encoding failed: %w", err)
			}

			targetEncoding, err = parseEncoding(flagTargetEncoding)
			if err != nil {
				return fmt.Errorf("parse target-encoding failed: %w", err)
			}

			if len(args) == 0 {
				args = []string{"-"}
				flagVerbose = false
			}

			if flagVerbose {
				logrus.SetLevel(logrus.DebugLevel)
			}

			for _, file := range args {
				logrus.Debugf("process %s start", file)
				if err := process(file); err != nil {
					logrus.WithError(err).Errorf("process %s error", file)
				}
				logrus.Debugf("process %s end", file)
			}

			return
		},
	}
)

func init() {
	cmd.Flags().SortFlags = false

	flags := cmd.PersistentFlags()
	{
		flags.SortFlags = false
		flags.BoolVarP(&flagReplace, "replace-source", "r", false, "replace/overwrite source file")
		flags.StringVarP(&flagSourceEncoding, "source-encoding", "s", "GBK", "source encoding")
		flags.StringVarP(&flagTargetEncoding, "target-encoding", "t", "UTF8", "target encoding")
		flags.BoolVarP(&flagVerbose, "verbose", "v", false, "verbose")
	}

	logrus.SetFormatter(&formatter.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		//NoColors:        true,
		HideKeys:    true,
		CallerFirst: true,
	})
}
func main() {
	_ = cmd.Execute()
}

func process(file string) (err error) {
	source, target, isDiskOp := os.Stdin, os.Stdout, file != "-"

	if isDiskOp {
		var abs string
		// resolve source path
		{
			logger := logrus.WithField("step", "RESOLVE")
			logger.Debugf("resolve %s", file)
			abs, err = filepath.Abs(file)
			if err != nil {
				return fmt.Errorf("cannot parse file path %s", file)
			}
			logger.Debugf("resolve %s done: %s", file, abs)
		}
		// open source file
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
		// create temp file
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

	// transfer source => target/temp file
	logger := logrus.WithField("step", "TRANSFER")
	logger.Debugf("transfer %s => %s", source.Name(), target.Name())
	_, err = io.Copy(
		transform.NewWriter(target, targetEncoding.NewEncoder()),
		transform.NewReader(source, sourceEncoding.NewDecoder()),
	)
	if err != nil {
		return fmt.Errorf("translate %s => %s failed: %s", source.Name(), target.Name(), err)
	}
	logger.Debugf("transfer %s => %s done", source.Name(), target.Name())

	// renaming target file
	if isDiskOp {
		saveTarget := source.Name() + ".out"

		if flagReplace {
			logger := logrus.WithField("step", "REMOVE")
			logger.Debugf("remove %s", source.Name())
			err = os.Remove(source.Name())
			if err != nil {
				return fmt.Errorf("remove %s failed: %w", source.Name(), err)
			}
			logger.Debugf("remove %s done", source.Name())

			saveTarget = source.Name()
		}

		logger := logrus.WithField("step", "RENAME")
		logger.Debugf("rename %s => %s", target.Name(), saveTarget)
		err = os.Rename(target.Name(), saveTarget)
		if err != nil {
			return fmt.Errorf("rename %s => %s failed: %w", target.Name(), saveTarget, err)
		}
		logger.Debugf("rename %s => %s done", target.Name(), saveTarget)

		logrus.Infof("save into %s", saveTarget)
	}

	return
}
func parseEncoding(encoding string) (enc encoding.Encoding, err error) {
	enc, err = htmlindex.Get(encoding)
	if err != nil {
		err = fmt.Errorf("invalid encoding: %s", encoding)
	}
	return
}
