package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

type options struct {
	SourceEncoding string   `short:"s" name:"source-encoding" default:"gbk" help:"Set source encoding."`
	TargetEncoding string   `short:"t" name:"target-encoding" default:"utf8" help:"Set target encoding."`
	Overwrite      bool     `short:"r" name:"overwrite" help:"Overwrite source file."`
	About          bool     `help:"Show about."`
	File           []string `arg:"" optional:""`
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
		c.File = append(c.File, "-")
	}

	for _, f := range c.File {
		err = c.process(f)
		if err != nil {
			return
		}
	}

	return
}
func (c *transcode) process(file string) (err error) {
	source, target := os.Stdin, os.Stdout

	if file != "-" {
		source, err = os.OpenFile(file, os.O_RDWR, 0755)
		if err != nil {
			return
		}
		defer source.Close()
		if c.Overwrite {
			target, err = os.CreateTemp(os.TempDir(), "")
			if err != nil {
				return
			}
			defer target.Close()
			defer func() {
				if err == nil {
					source.Seek(0, io.SeekStart)
					target.Seek(0, io.SeekStart)
					_, err = io.Copy(source, target)
				}
			}()
		}
	}

	_, err = io.Copy(
		transform.NewWriter(target, c.target.NewEncoder()),
		transform.NewReader(source, c.source.NewDecoder()),
	)

	return
}

func parseEncoding(encoding string) (enc encoding.Encoding, err error) {
	enc, err = htmlindex.Get(encoding)
	if err != nil {
		err = fmt.Errorf("invalid encoding: %s", encoding)
	}
	return
}
