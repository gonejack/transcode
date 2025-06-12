package chardet

import (
	"fmt"

	"github.com/gogs/chardet"
)

func DetectEncodingByGogsChardet(dat []byte) (string, error) {
	v, err := chardet.NewTextDetector().DetectBest(dat)
	if err != nil {
		return "", fmt.Errorf("detect failed by github.com/wlynxg/chardet: %s", err)
	}
	return v.Charset, nil
}
