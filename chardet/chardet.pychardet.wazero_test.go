//go:build pychardet || pychardet_wazero

package chardet

import "testing"

func TestNormalizePyChardetEncoding(t *testing.T) {
	cases := map[string]string{
		"GB2312":       "gb2312",
		"IBM866":       "cp866",
		"SHIFT_JIS":    "shift-jis",
		"UTF-8":        "utf-8",
		"UTF-8-SIG":    UTF8WithBOM,
		"MacRoman":     "macintosh",
		"MacCyrillic":  "x-mac-cyrillic",
		"Windows-1252": "windows-1252",
		"KOI8-R":       "koi8-r",
		"  EUC-JP  ":   "euc-jp",
	}
	for in, want := range cases {
		if got := normalizePyChardetEncoding(in); got != want {
			t.Errorf("normalizePyChardetEncoding(%q) = %q, want %q", in, got, want)
		}
	}
}
