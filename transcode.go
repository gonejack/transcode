package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/gogs/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type options struct {
	SourceEncoding string   `short:"s" name:"source-encoding" default:"auto" help:"Set source encoding, default as auto-detection."`
	TargetEncoding string   `short:"t" name:"target-encoding" default:"utf8" help:"Set target encoding, default as utf8."`
	DetectEncoding bool     `short:"d" name:"detect-encoding" help:"Detect encoding only."`
	Overwrite      bool     `short:"w" name:"overwrite" help:"Overwrite source file."`
	About          bool     `help:"Show about."`
	File           []string `arg:"" optional:""`
}
type trans struct {
	options
	source encoding.Encoding
	target encoding.Encoding
}

func (c *trans) run() (err error) {
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
		err = c.proc(f)
		if err != nil {
			return fmt.Errorf("process %s failed: %w", f, err)
		}
	}
	return
}
func (c *trans) proc(f string) (err error) {
	src, out := os.Stdin, os.Stdout
	if f != "-" {
		if c.Overwrite {
			src, err = os.OpenFile(f, os.O_RDWR, 0)
		} else {
			src, err = os.Open(f)
		}
		if err != nil {
			return
		}
		defer src.Close()
		st, exx := src.Stat()
		switch {
		case exx != nil:
			return fmt.Errorf("read file info failed: %w", exx)
		case !st.Mode().IsRegular():
			return errors.New("not a regular file")
		case st.Size() == 0:
			log.Printf("no changes, source file %s is empty", f)
			return
		}
	}
	srd := bufio.NewReader(src)
	switch {
	case c.DetectEncoding:
		enc, exx := detectEncoding(srd)
		if exx != nil {
			fmt.Printf("detect encoding of %s failed: %s", f, exx)
		} else {
			fmt.Printf("detected encoding of %s is %s", f, enc)
		}
		return
	case strings.EqualFold(c.SourceEncoding, "auto"):
		c.source, err = autoEncoding(srd)
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
			log.Printf("no changes, source file %s is already in target encoding %s", f, c.target)
			return
		}
		out, err = os.CreateTemp(os.TempDir(), "transcode.*.txt")
		if err != nil {
			return
		}
		defer func() {
			if err == nil {
				src.Truncate(0)
				src.Seek(0, io.SeekStart)
				out.Seek(0, io.SeekStart)
				_, err = io.Copy(src, out)
			}
			out.Close()
			os.Remove(out.Name())
		}()
	}
	_, err = io.Copy(transform.NewWriter(out, c.target.NewEncoder()), transform.NewReader(srd, c.source.NewDecoder()))
	return
}

func autoEncoding(r *bufio.Reader) (enc encoding.Encoding, err error) {
	coding, err := detectEncoding(r)
	if err == nil {
		enc, err = parseEncoding(coding)
	}
	return
}
func parseEncoding(encoding string) (enc encoding.Encoding, err error) {
	enc, err = htmlindex.Get(encoding)
	if err != nil {
		err = fmt.Errorf("invalid encoding: %s", encoding)
	}
	switch enc {
	case simplifiedchinese.GBK:
		enc = simplifiedchinese.GB18030
	}
	return
}
func detectEncoding(r *bufio.Reader) (string, error) {
	hdr, err := r.Peek(2048)
	if len(hdr) == 0 {
		return "", fmt.Errorf("cannot read file: %w", err)
	}
	res, err := chardet.NewTextDetector().DetectBest(hdr)
	if err != nil {
		return "", err
	}
	return res.Charset, nil
}
