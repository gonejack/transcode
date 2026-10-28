//go:build charamel_wazero || pychardet || pychardet_wazero

package chardet

import "embed"

const embeddedPythonWasiRoot = "python_wasi"

// The CPython WASI interpreter and reduced standard library are shared by
// every Python-based detector. Keeping these embeds in one file avoids
// embedding the same 30+ MiB wasm binary twice when multiple Python detectors
// are compiled into the same binary.
//
//go:embed python_wasi/python.wasm
var embeddedPythonWasi []byte

//go:embed all:python_wasi/lib
var embeddedPythonWasiLib embed.FS
