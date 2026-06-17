package chardet

import (
	"os"
	"testing"
)

var cases = []struct {
	file string
	want string
}{
	{"../testfiles/野草utf8.txt", "utf-8"},
	{"../testfiles/test.txt", UTF16LEWithBOM},
	{"../testfiles/abc.txt", "gb2312"},
	{"../testfiles/gb18030.txt", "gb2312"},
	{"../testfiles/hello.txt", "cp866"},
}

func TestDetectEncoding(t *testing.T) {
	for _, c := range cases {
		dat, err := os.ReadFile(c.file)
		if err != nil {
			t.Fatalf("read %s: %v", c.file, err)
		}
		got, err := DetectEncoding(dat)
		if err != nil {
			t.Errorf("%s: error %v", c.file, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.file, got, c.want)
		}
	}
}
