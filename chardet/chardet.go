package chardet

import (
	gogschardet "github.com/gogs/chardet"
	"github.com/wlynxg/chardet"
)

func detectEncoding(dat []byte) (string, error) {
	r1 := chardet.Detect(dat)
	if r1.Encoding > "" {
		return r1.Encoding, nil
	}
	r2, err := gogschardet.NewTextDetector().DetectBest(dat)
	if err != nil {
		return "", err
	}
	return r2.Charset, nil
}
