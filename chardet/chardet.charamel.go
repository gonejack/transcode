//go:build charamel

package chardet

import (
	"errors"

	"github.com/gonejack/charamel"
)

func init() {
	prefer(DetectEncodingByCharamel)
}

var encodings = []charamel.Encoding{
	// UTF编码
	charamel.UTF8,
	charamel.UTF16,
	charamel.UTF16BE,
	charamel.UTF16LE,
	charamel.UTF32,
	charamel.UTF32BE,
	charamel.UTF32LE,

	// ASCII
	charamel.ASCII,

	// 西欧编码 (ISO-8859)
	charamel.LATIN1,    // ISO-8859-1
	charamel.ISO88592,  // ISO-8859-2
	charamel.ISO88593,  // ISO-8859-3
	charamel.ISO88594,  // ISO-8859-4
	charamel.ISO88595,  // ISO-8859-5
	charamel.ISO88596,  // ISO-8859-6
	charamel.ISO88597,  // ISO-8859-7
	charamel.ISO88598,  // ISO-8859-8
	charamel.ISO88599,  // ISO-8859-9
	charamel.ISO885910, // ISO-8859-10
	charamel.ISO885911, // ISO-8859-11
	charamel.ISO885913, // ISO-8859-13
	charamel.ISO885914, // ISO-8859-14
	charamel.ISO885915, // ISO-8859-15
	charamel.ISO885916, // ISO-8859-16

	// Windows编码 (CP1250-1258)
	charamel.CP1250,
	charamel.CP1251,
	charamel.CP1252,
	charamel.CP1253,
	charamel.CP1254,
	charamel.CP1255,
	charamel.CP1256,
	charamel.CP1257,
	charamel.CP1258,

	// 中文编码
	charamel.GB2312,
	charamel.GBK,
	charamel.GB18030,
	charamel.BIG5,
	charamel.BIG5HKSCS,
	charamel.HZ, // hz-gb-2312

	// 日文编码
	charamel.EUCJP,     // euc-jp
	charamel.SHIFTJIS,  // shift_jis
	charamel.ISO2022JP, // iso-2022-jp

	// 韩文编码
	charamel.EUCKR,     // euc-kr
	charamel.CP949,     // windows-949
	charamel.ISO2022KR, // iso-2022-kr

	// 俄文编码
	charamel.KOI8R, // koi8-r
	charamel.KOI8U, // koi8-u
	charamel.CP866, // cp866

	// 泰文编码
	charamel.TIS620, // tis-620
	charamel.CP874,  // windows-874

	// Mac编码
	charamel.MACROMAN,    // macintosh
	charamel.MACCYRILLIC, // x-mac-cyrillic
}

func DetectEncodingByCharamel(dat []byte) (string, error) {
	d, err := charamel.NewDetector(encodings, 0)
	if err != nil {
		return "", err
	}
	v := d.Detect(dat)
	if v == nil {
		return "", errors.New("detect failed by github.com/gonejack/charamel")
	}
	return mapName(v), nil
}

func mapName(encoding *charamel.Encoding) string {
	switch *encoding {
	// UTF编码
	case charamel.UTF8:
		return "utf-8"
	case charamel.UTF16:
		return "utf-16"
	case charamel.UTF16BE:
		return "utf-16be"
	case charamel.UTF16LE:
		return "utf-16le"
	case charamel.UTF32:
		return "utf-32"
	case charamel.UTF32BE:
		return "utf-32be"
	case charamel.UTF32LE:
		return "utf-32le"

	// ASCII
	case charamel.ASCII:
		return "ascii"

	// 西欧编码 (ISO-8859)
	case charamel.LATIN1:
		return "iso-8859-1"
	case charamel.ISO88592:
		return "iso-8859-2"
	case charamel.ISO88593:
		return "iso-8859-3"
	case charamel.ISO88594:
		return "iso-8859-4"
	case charamel.ISO88595:
		return "iso-8859-5"
	case charamel.ISO88596:
		return "iso-8859-6"
	case charamel.ISO88597:
		return "iso-8859-7"
	case charamel.ISO88598:
		return "iso-8859-8"
	case charamel.ISO88599:
		return "iso-8859-9"
	case charamel.ISO885910:
		return "iso-8859-10"
	case charamel.ISO885911:
		return "iso-8859-11"
	case charamel.ISO885913:
		return "iso-8859-13"
	case charamel.ISO885914:
		return "iso-8859-14"
	case charamel.ISO885915:
		return "iso-8859-15"
	case charamel.ISO885916:
		return "iso-8859-16"

	// Windows编码
	case charamel.CP1250:
		return "cp1250"
	case charamel.CP1251:
		return "cp1251"
	case charamel.CP1252:
		return "cp1252"
	case charamel.CP1253:
		return "cp1253"
	case charamel.CP1254:
		return "cp1254"
	case charamel.CP1255:
		return "cp1255"
	case charamel.CP1256:
		return "cp1256"
	case charamel.CP1257:
		return "cp1257"
	case charamel.CP1258:
		return "cp1258"

	// 中文编码
	case charamel.GB2312:
		return "gb2312"
	case charamel.GBK:
		return "gbk"
	case charamel.GB18030:
		return "gb18030"
	case charamel.BIG5:
		return "big5"
	case charamel.BIG5HKSCS:
		return "big5-hkscs"
	case charamel.HZ:
		return "hz-gb-2312"

	// 日文编码
	case charamel.EUCJP:
		return "euc-jp"
	case charamel.SHIFTJIS:
		return "shift-jis"
	case charamel.ISO2022JP:
		return "iso-2022-jp"

	// 韩文编码
	case charamel.EUCKR:
		return "euc-kr"
	case charamel.CP949:
		return "windows-949"
	case charamel.ISO2022KR:
		return "iso-2022-kr"

	// 俄文编码
	case charamel.KOI8R:
		return "koi8-r"
	case charamel.KOI8U:
		return "koi8-u"
	case charamel.CP866:
		return "cp866"

	// 泰文编码
	case charamel.TIS620:
		return "tis-620"
	case charamel.CP874:
		return "windows-874"

	// Mac编码
	case charamel.MACROMAN:
		return "macintosh"
	case charamel.MACCYRILLIC:
		return "x-mac-cyrillic"

	// 默认返回原始字符串
	default:
		return encoding.String()
	}
}
