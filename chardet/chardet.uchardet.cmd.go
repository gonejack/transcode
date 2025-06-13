package chardet

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func DetectEncodingByUChardetCmd(dat []byte) (string, error) {
	name, err := lookupUchardet()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(name)
	cmd.Stdin = bytes.NewReader(dat)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute uchardet command: %w", err)
	}
	v := strings.TrimSpace(string(out))
	if v == "" || v == "unknown" {
		return "", errors.New("detect failed by uchardet command")
	}
	return v, nil
}

func lookupUchardet() (string, error) {
	var ns = []string{"uchardet"}
	if runtime.GOOS == "windows" {
		ns = []string{"./uchardet.exe", "uchardet.exe", "uchardet"}
	}
	for _, name := range ns {
		_, err := exec.LookPath(name)
		if err == nil {
			return name, nil
		}
	}
	return "", errors.New("command uchardet not found in system PATH")
}
