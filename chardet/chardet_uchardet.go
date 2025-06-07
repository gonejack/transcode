//go:build uchardet

package chardet

import (
	"github.com/gonejack/transcode/chardet/uchardet"
)

func DetectEncoding(dat []byte) (string, error) {
	dec := uchardet.NewChardet()
	defer dec.Release()
	if dec.Handle(dat) == 0 {
		if v := dec.End(); v > "" {
			return v, nil
		}
	}
	return detectEncoding(dat)
}
