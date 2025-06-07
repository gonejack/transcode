package chardet

func DetectEncoding(dat []byte) (string, error) {
	return detectEncoding(dat)
}
