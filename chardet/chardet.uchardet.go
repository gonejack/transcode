//go:build uchardet

package chardet

import (
	"errors"

	"github.com/gonejack/transcode/chardet/uchardet"
)

func init() {
	prefer(DetectEncodingByUChardet)
}

func DetectEncodingByUChardet(dat []byte) (string, error) {
	dec := uchardet.NewChardet()
	defer dec.Release()
	if dec.Handle(dat) == 0 {
		if v := dec.End(); v > "" {
			return v, nil
		}
	}
	return "", errors.New("detect failed by uchardet")
}
