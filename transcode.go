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
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/gonejack/transcode/chardet"
)

type options struct {
	SourceEncoding string   `short:"s" name:"source-encoding" default:"auto" help:"Set source encoding, default as auto-detection."`
	TargetEncoding string   `short:"t" name:"target-encoding" default:"utf8" help:"Set target encoding, default as utf8."`
	DetectEncoding bool     `short:"d" name:"detect-encoding" help:"Detect encoding only."`
	Overwrite      bool     `short:"w" name:"overwrite" help:"Overwrite source file."`
	ListEncodings  bool     `short:"l" name:"list-encodings" help:"list supported encodings"`
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
	if c.ListEncodings {
		fmt.Println("Supported encodings:")
		fmt.Println(strings.Join(encodings(), "\n"))
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
			fmt.Printf("detecting encoding of file %s failed: %s", f, exx)
		} else {
			fmt.Printf("encoding of file %s is %s", f, enc)
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
	r := transform.NewReader(srd, c.source.NewDecoder())
	w := transform.NewWriter(out, c.target.NewEncoder())
	_, err = io.Copy(w, r)
	w.Close()
	return
}

func autoEncoding(r *bufio.Reader) (enc encoding.Encoding, err error) {
	coding, err := detectEncoding(r)
	if err == nil {
		enc, err = parseEncoding(coding)
	}
	return
}
func detectEncoding(r *bufio.Reader) (string, error) {
	hdr, err := r.Peek(2048)
	if len(hdr) == 0 {
		return "", fmt.Errorf("cannot read input data: %w", err)
	}
	return chardet.DetectEncoding(hdr)
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
func encodings() []string {
	list := []string{
		"unicode-1-1-utf-8",
		"unicode11utf8",
		"unicode20utf8",
		"utf-8",
		"utf8",
		"x-unicode20utf8",
		"866",
		"cp866",
		"csibm866",
		"ibm866",
		"csisolatin2",
		"iso-8859-2",
		"iso-ir-101",
		"iso8859-2",
		"iso88592",
		"iso_8859-2",
		"iso_8859-2:1987",
		"l2",
		"latin2",
		"csisolatin3",
		"iso-8859-3",
		"iso-ir-109",
		"iso8859-3",
		"iso88593",
		"iso_8859-3",
		"iso_8859-3:1988",
		"l3",
		"latin3",
		"csisolatin4",
		"iso-8859-4",
		"iso-ir-110",
		"iso8859-4",
		"iso88594",
		"iso_8859-4",
		"iso_8859-4:1988",
		"l4",
		"latin4",
		"csisolatincyrillic",
		"cyrillic",
		"iso-8859-5",
		"iso-ir-144",
		"iso8859-5",
		"iso88595",
		"iso_8859-5",
		"iso_8859-5:1988",
		"arabic",
		"asmo-708",
		"csiso88596e",
		"csiso88596i",
		"csisolatinarabic",
		"ecma-114",
		"iso-8859-6",
		"iso-8859-6-e",
		"iso-8859-6-i",
		"iso-ir-127",
		"iso8859-6",
		"iso88596",
		"iso_8859-6",
		"iso_8859-6:1987",
		"csisolatingreek",
		"ecma-118",
		"elot_928",
		"greek",
		"greek8",
		"iso-8859-7",
		"iso-ir-126",
		"iso8859-7",
		"iso88597",
		"iso_8859-7",
		"iso_8859-7:1987",
		"sun_eu_greek",
		"csiso88598e",
		"csisolatinhebrew",
		"hebrew",
		"iso-8859-8",
		"iso-8859-8-e",
		"iso-ir-138",
		"iso8859-8",
		"iso88598",
		"iso_8859-8",
		"iso_8859-8:1988",
		"visual",
		"csiso88598i",
		"iso-8859-8-i",
		"logical",
		"csisolatin6",
		"iso-8859-10",
		"iso-ir-157",
		"iso8859-10",
		"iso885910",
		"l6",
		"latin6",
		"iso-8859-13",
		"iso8859-13",
		"iso885913",
		"iso-8859-14",
		"iso8859-14",
		"iso885914",
		"csisolatin9",
		"iso-8859-15",
		"iso8859-15",
		"iso885915",
		"iso_8859-15",
		"l9",
		"iso-8859-16",
		"cskoi8r",
		"koi",
		"koi8",
		"koi8-r",
		"koi8_r",
		"koi8-ru",
		"koi8-u",
		"csmacintosh",
		"mac",
		"macintosh",
		"x-mac-roman",
		"dos-874",
		"iso-8859-11",
		"iso8859-11",
		"iso885911",
		"tis-620",
		"windows-874",
		"cp1250",
		"windows-1250",
		"x-cp1250",
		"cp1251",
		"windows-1251",
		"x-cp1251",
		"ansi_x3.4-1968",
		"ascii",
		"cp1252",
		"cp819",
		"csisolatin1",
		"ibm819",
		"iso-8859-1",
		"iso-ir-100",
		"iso8859-1",
		"iso88591",
		"iso_8859-1",
		"iso_8859-1:1987",
		"l1",
		"latin1",
		"us-ascii",
		"windows-1252",
		"x-cp1252",
		"cp1253",
		"windows-1253",
		"x-cp1253",
		"cp1254",
		"csisolatin5",
		"iso-8859-9",
		"iso-ir-148",
		"iso8859-9",
		"iso88599",
		"iso_8859-9",
		"iso_8859-9:1989",
		"l5",
		"latin5",
		"windows-1254",
		"x-cp1254",
		"cp1255",
		"windows-1255",
		"x-cp1255",
		"cp1256",
		"windows-1256",
		"x-cp1256",
		"cp1257",
		"windows-1257",
		"x-cp1257",
		"cp1258",
		"windows-1258",
		"x-cp1258",
		"x-mac-cyrillic",
		"x-mac-ukrainian",
		"chinese",
		"csgb2312",
		"csiso58gb231280",
		"gb2312",
		"gb_2312",
		"gb_2312-80",
		"gbk",
		"iso-ir-58",
		"x-gbk",
		"gb18030",
		"big5",
		"big5-hkscs",
		"cn-big5",
		"csbig5",
		"x-x-big5",
		"cseucpkdfmtjapanese",
		"euc-jp",
		"x-euc-jp",
		"csiso2022jp",
		"iso-2022-jp",
		"csshiftjis",
		"ms932",
		"ms_kanji",
		"shift-jis",
		"shift_jis",
		"sjis",
		"windows-31j",
		"x-sjis",
		"cseuckr",
		"csksc56011987",
		"euc-kr",
		"iso-ir-149",
		"korean",
		"ks_c_5601-1987",
		"ks_c_5601-1989",
		"ksc5601",
		"ksc_5601",
		"windows-949",
		"csiso2022kr",
		"hz-gb-2312",
		"iso-2022-cn",
		"iso-2022-cn-ext",
		"iso-2022-kr",
		"replacement",
		"unicodefffe",
		"utf-16be",
		"csunicode",
		"iso-10646-ucs-2",
		"ucs-2",
		"unicode",
		"unicodefeff",
		"utf-16",
		"utf-16le",
		"x-user-defined",
	}
	return list
}
