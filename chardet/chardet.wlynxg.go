package chardet

import (
	"errors"

	"github.com/wlynxg/chardet"
)

func DetectEncodingByWlynxgChardet(dat []byte) (string, error) {
	v := chardet.Detect(dat)
	if v.Encoding > "" {
		return v.Encoding, nil
	}
	return "", errors.New("detect failed by github.com/wlynxg/chardet")
}
