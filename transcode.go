package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/gogs/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

type options struct {
	SourceEncoding string   `short:"s" name:"source-encoding" default:"auto" help:"Set source encoding, default as auto-detection."`
	TargetEncoding string   `short:"t" name:"target-encoding" default:"utf8" help:"Set target encoding, default as utf8."`
	Overwrite      bool     `short:"w" name:"overwrite" help:"Overwrite source file."`
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

	if len(c.File) == 0 {
		c.File = append(c.File, "-")
	}

	c.target, err = parseEncoding(c.TargetEncoding)
	if err != nil {
		return fmt.Errorf("parse target-encoding %s failed: %w", c.TargetEncoding, err)
	}

	for _, f := range c.File {
		err = c.process(f)
		if err != nil {
			return fmt.Errorf("process %s failed: %w", f, err)
		}
	}

	return
}
func (c *transcode) process(file string) (err error) {
	src, dst := os.Stdin, os.Stdout
	if file != "-" {
		src, err = os.OpenFile(file, os.O_RDWR, 0755)
		if err != nil {
			return
		}
		defer src.Close()
		stat, exx := src.Stat()
		if exx != nil {
			return fmt.Errorf("read file info failed: %w", exx)
		}
		if stat.Size() == 0 {
			log.Printf("no changes, source file %s is empty", file)
			return
		}
	}
	srd := bufio.NewReader(src)
	switch {
	case strings.EqualFold(c.SourceEncoding, "auto"):
		c.source, err = detectEncoding(srd)
		if err != nil {
			return fmt.Errorf("cannot determine source-encoding: %w", err)
		}
	default:
		c.source, err = parseEncoding(c.SourceEncoding)
		if err != nil {
			return fmt.Errorf("parse source-encoding %s failed: %w", c.SourceEncoding, err)
		}
	}
	if src != os.Stdin && c.Overwrite {
		if c.source == c.target {
			log.Printf("no changes, source file %s is already in target encoding %s", file, c.target)
			return
		}
		dst, err = os.CreateTemp(os.TempDir(), "")
		if err != nil {
			return
		}
		defer os.Remove(dst.Name())
		defer dst.Close()
		defer func() {
			if err == nil {
				src.Truncate(0)
				src.Seek(0, io.SeekStart)
				dst.Seek(0, io.SeekStart)
				_, err = io.Copy(src, dst)
			}
		}()
	}
	_, err = io.Copy(
		transform.NewWriter(dst, c.target.NewEncoder()),
		transform.NewReader(srd, c.source.NewDecoder()),
	)
	return
}

func detectEncoding(r *bufio.Reader) (e encoding.Encoding, err error) {
	dat, err := r.Peek(2048)
	if len(dat) == 0 {
		return nil, fmt.Errorf("cannot detect encoding: %w", err)
	}
	res, err := chardet.NewTextDetector().DetectBest(dat)
	if err != nil {
		return
	}
	return parseEncoding(res.Charset)
}
func parseEncoding(encoding string) (enc encoding.Encoding, err error) {
	enc, err = htmlindex.Get(encoding)
	if err != nil {
		err = fmt.Errorf("invalid encoding: %s", encoding)
	}
	return
}
