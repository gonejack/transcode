package uchardet

// https://github.com/centny/uchardet

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <uchardet/uchardet.h>
// Apple Silicon (M1/M2/M3)
#cgo darwin,arm64 CPPFLAGS: -I/opt/homebrew/opt/uchardet/include
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/uchardet/lib -luchardet

// Intel macOS (x86_64)
#cgo darwin,amd64 CPPFLAGS: -I/usr/local/include
#cgo darwin,amd64 LDFLAGS: -L/usr/local/lib -luchardet

// Linux (assumes installed in /usr/local)
#cgo linux CPPFLAGS: -I/usr/local/include
#cgo linux LDFLAGS: -L/usr/local/lib -luchardet
*/
import "C"
import "unsafe"

// Chardet is the binding uchardet_t on libuchardet
type Chardet struct {
	det C.uchardet_t
}

// NewChardet is the default creator to create Chardet
func NewChardet() *Chardet {
	return &Chardet{
		det: C.uchardet_new(),
	}
}

// Release will free the Chardet
func (c *Chardet) Release() {
	C.uchardet_delete(c.det)
}

// Handle will process the data slice
func (c *Chardet) Handle(buf []byte) int {
	var data = (*C.char)(unsafe.Pointer(&buf[0]))
	var dlen = C.size_t(len(buf))
	return int(C.uchardet_handle_data(c.det, data, dlen))
}

// Reset encoding detector.
func (c *Chardet) Reset() {
	C.uchardet_reset(c.det)
}

// End is ending the process and return the encoding name
func (c *Chardet) End() string {
	C.uchardet_data_end(c.det)
	return cstring(C.uchardet_get_charset(c.det))
}

func cstring(cs *C.char) string {
	clen := C.strlen(cs)
	if clen < 1 {
		return ""
	}
	buf := make([]byte, clen+1)
	C.strcpy((*C.char)(unsafe.Pointer(&buf[0])), cs)
	return string(buf[:clen])
}
