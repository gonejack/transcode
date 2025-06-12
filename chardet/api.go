package chardet

import "slices"

type detectFunc func([]byte) (string, error)

func prefer(f detectFunc) {
	detectFuncList = slices.Insert(detectFuncList, 0, f)
}

var detectFuncList = []detectFunc{
	DetectEncodingByWlynxgChardet,
	DetectEncodingByGogsChardet,
}

func DetectEncoding(dat []byte) (v string, err error) {
	for _, df := range detectFuncList {
		v, err = df(dat)
		if err == nil {
			break
		}
	}
	return
}
