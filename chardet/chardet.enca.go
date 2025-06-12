//go:build enca

package chardet

import (
	"github.com/endeveit/enca"
)

func init() {
	prefer(DetectEncodingByEnca)
}

func DetectEncodingByEnca(dat []byte) (string, error) {
	ana, err := enca.New("__")
	if err != nil {
		return "", err
	}
	defer ana.Free()
	v, err := ana.FromBytes(dat, enca.NAME_STYLE_RFC1345)
	return v, err
}
