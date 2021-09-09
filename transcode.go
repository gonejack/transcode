package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

type options struct {
	SourceEncoding string `short:"s" name:"source-encoding" default:"gbk" help:"Set source encoding."`
	TargetEncoding string `short:"t" name:"target-encoding" default:"utf8" help:"Set target encoding."`
	Overwrite      bool   `short:"r" name:"overwrite" help:"Overwrite source file."`
	Verbose        bool   `short:"v" help:"Verbose printing."`
	About          bool   `help:"Show about."`

	File []string `arg:"" optional:""`
}
type transcode struct {
	options
	source encoding.Encoding
	target encoding.Encoding
}

func (c *transcode) run() (err error) {
	kong.Parse(&c.options,
		kong.Name("transcode"),
		kong.Description("Translate text encoding."),
		kong.UsageOnError(),
	)

	if c.About {
		fmt.Println("Visit https://github.com/gonejack/transcode")
		return
	}

	c.source, err = parseEncoding(c.SourceEncoding)
	if err != nil {
		return fmt.Errorf("parse source-encoding failed: %w", err)
	}
	c.target, err = parseEncoding(c.TargetEncoding)
	if err != nil {
		return fmt.Errorf("parse target-encoding failed: %w", err)
	}

	if len(c.File) == 0 {
		c.File = []string{"-"}
		c.Verbose = false
	}

	if c.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	for _, file := range c.File {
		logrus.Debugf("process %s start", file)
		err := c.process(file)
		if err != nil {
			logrus.WithError(err).Errorf("process %s error", file)
		}
		logrus.Debugf("process %s end", file)
	}

	return
}

func (c *transcode) process(file string) (err error) {
	source, target, writeDisk := os.Stdin, os.Stdout, file != "-"

	if writeDisk {
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
		transform.NewWriter(target, c.target.NewEncoder()),
		transform.NewReader(source, c.source.NewDecoder()),
	)
	if err != nil {
		return fmt.Errorf("translate %s => %s failed: %s", source.Name(), target.Name(), err)
	}
	logger.Debugf("transfer %s => %s done", source.Name(), target.Name())

	// renaming target file
	if writeDisk {
		output := source.Name() + ".out"

		if c.Overwrite {
			logger := logrus.WithField("step", "REMOVE")
			logger.Debugf("remove %s", source.Name())
			err = os.Remove(source.Name())
			if err != nil {
				return fmt.Errorf("remove %s failed: %w", source.Name(), err)
			}
			logger.Debugf("remove %s done", source.Name())
			output = source.Name()
		}

		logger := logrus.WithField("step", "RENAME")
		logger.Debugf("rename %s => %s", target.Name(), output)
		err = os.Rename(target.Name(), output)
		if err != nil {
			return fmt.Errorf("rename %s => %s failed: %w", target.Name(), output, err)
		}
		logger.Debugf("rename %s => %s done", target.Name(), output)
		logrus.Infof("save into %s", output)
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
