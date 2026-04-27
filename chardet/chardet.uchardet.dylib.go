//go:build unix

package chardet

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

func DetectEncodingByUChardetDylib(dat []byte) (string, error) {
	if lib == 0 {
		return "", errors.New("no uchardet dylib found")
	}
	dec := NewChardet()
	defer dec.Release()
	if dec.Handle(dat) == 0 {
		if v := dec.End(); v > "" {
			return v, nil
		}
	}
	return "", errors.New("detect failed by uchardet")
}

var (
	lib                uintptr
	uchardetNew        func() uintptr
	uchardetDelete     func(det uintptr)
	uchardetHandleData func(det uintptr, data []byte, len int) int
	uchardetDataEnd    func(det uintptr)
	uchardetGetCharset func(det uintptr) *byte // 返回 C 字符串 (char*)
	uchardetReset      func(det uintptr)
)

func init() {
	var err error
	var name = uchardetLib()
	lib, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return
	}
	purego.RegisterLibFunc(&uchardetNew, lib, "uchardet_new")
	purego.RegisterLibFunc(&uchardetDelete, lib, "uchardet_delete")
	purego.RegisterLibFunc(&uchardetHandleData, lib, "uchardet_handle_data")
	purego.RegisterLibFunc(&uchardetDataEnd, lib, "uchardet_data_end")
	purego.RegisterLibFunc(&uchardetGetCharset, lib, "uchardet_get_charset")
	purego.RegisterLibFunc(&uchardetReset, lib, "uchardet_reset")
}

type Chardet struct {
	det uintptr
}

func NewChardet() *Chardet {
	if uchardetNew == nil {
		return nil
	}
	return &Chardet{
		det: uchardetNew(),
	}
}
func (c *Chardet) Release() {
	if c.det != 0 {
		uchardetDelete(c.det)
		c.det = 0
	}
}
func (c *Chardet) Handle(buf []byte) int {
	if c.det == 0 || len(buf) == 0 {
		return -1 // 或其他错误指示
	}
	dlen := len(buf)
	return uchardetHandleData(c.det, buf, dlen)
}
func (c *Chardet) End() string {
	uchardetDataEnd(c.det)
	cString := uchardetGetCharset(c.det)
	return cstrToString(cString)
}

func uchardetLib() string {
	switch runtime.GOOS {
	case "darwin":
		return "/opt/homebrew/lib/libuchardet.dylib"
	case "linux":
		return "libuchardet.so"
	default:
		return ""
	}
}
func cstrToSlice(cptr *byte) []byte {
	if cptr == nil {
		return nil
	}
	var length int
	for ptr := cptr; *ptr != 0; ptr = (*byte)(unsafe.Add(unsafe.Pointer(ptr), 1)) {
		length++
	}
	return unsafe.Slice(cptr, length)
}
func cstrToString(cptr *byte) string {
	return string(cstrToSlice(cptr))
}
