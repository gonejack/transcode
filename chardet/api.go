package chardet

import (
	"bytes"
	"slices"

	"github.com/wlynxg/chardet/consts"
)

type detectFunc func([]byte) (string, error)

func prefer(f detectFunc) {
	detectFuncList = slices.Insert(detectFuncList, 0, f)
}

var detectFuncList = []detectFunc{
	//DetectEncodingByUChardetDylib,
	DetectEncodingByUChardetCmd,
	DetectEncodingByWlynxgChardet,
	DetectEncodingByGogsChardet,
}

const (
	UTF8WithBOM    string = "utf-8-bom"
	UTF16LEWithBOM string = "utf-16le-bom"
	UTF16BEWithBOM string = "utf-16be-bom"
	UTF32LEWithBOM string = "utf-32le-bom"
	UTF32BEWithBOM string = "utf-32be-bom"
)

func DetectEncoding(dat []byte) (v string, err error) {
	switch {
	case bytes.HasPrefix(dat, []byte(consts.UTF8BOM)):
		return UTF8WithBOM, nil // EF BB BF  UTF-8 with BOM
	case bytes.HasPrefix(dat, []byte(consts.UTF16LEBOM)):
		return UTF16LEWithBOM, nil // FF FE  UTF-16, little endian BOM
	case bytes.HasPrefix(dat, []byte(consts.UTF16BEBOM)):
		return UTF16BEWithBOM, nil // FE FF  UTF-16, big endian BOM
	case bytes.HasPrefix(dat, []byte(consts.UTF32LEBOM)):
		return UTF32LEWithBOM, nil // FF FE 00 00  UTF-32, little-endian BOM
	case bytes.HasPrefix(dat, []byte(consts.UTF32BEBOM)):
		return UTF32BEWithBOM, nil // 00 00 FE FF  UTF-32, big-endian BOM
	}
	for _, f := range detectFuncList {
		v, err = f(dat)
		if err == nil {
			break
		}
	}
	return
}
