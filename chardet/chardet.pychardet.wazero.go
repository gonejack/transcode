//go:build pychardet || pychardet_wazero

package chardet

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

//go:embed all:pychardet_wasi/app
var embeddedPyChardetWasiApp embed.FS

const embeddedPyChardetWasiRoot = "pychardet_wasi"

type pyChardetRuntimeState struct {
	once     sync.Once
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
	rootFS   fs.FS
	appFS    fs.FS
	err      error
}

var pyChardetRuntime pyChardetRuntimeState

func init() {
	prefer(DetectEncodingByPyChardet)
}

// DetectEncodingByPyChardet detects the encoding using the Python chardet
// package running on the embedded CPython WASI runtime.
func DetectEncodingByPyChardet(dat []byte) (string, error) {
	if len(dat) == 0 {
		return "", errors.New("pychardet wasm: empty input")
	}

	ctx := context.Background()
	if err := initPyChardetWasmRuntime(ctx); err != nil {
		return "", err
	}

	var stdout, stderr bytes.Buffer
	mod, err := pyChardetRuntime.runtime.InstantiateModule(
		ctx,
		pyChardetRuntime.compiled,
		wazero.NewModuleConfig().
			WithName("").
			WithArgs("python", "-S", "-B", "/app/detect.py").
			WithEnv("PYTHONPATH", "/app").
			WithEnv("PYTHONHOME", "/").
			WithEnv("PYTHONDONTWRITEBYTECODE", "1").
			WithStdin(bytes.NewReader(dat)).
			WithStdout(&stdout).
			WithStderr(&stderr).
			WithFSConfig(pyChardetRuntime.fsConfig()).
			WithSysWalltime().
			WithSysNanotime().
			WithSysNanosleep().
			WithRandSource(rand.Reader),
	)
	if mod != nil {
		defer mod.Close(ctx)
	}
	if err != nil {
		return "", formatPyChardetWasmError("run", err, stderr.String())
	}

	result := normalizePyChardetEncoding(stdout.String())
	if result == "" {
		return "", formatPyChardetWasmError("run", errors.New("empty result"), stderr.String())
	}
	return result, nil
}

func initPyChardetWasmRuntime(ctx context.Context) error {
	pyChardetRuntime.once.Do(func() {
		rootFS, err := fs.Sub(embeddedPythonWasiLib, embeddedPythonWasiRoot)
		if err != nil {
			pyChardetRuntime.err = fmt.Errorf("pychardet wasm embedded runtime unavailable: %w", err)
			return
		}
		appFS, err := fs.Sub(embeddedPyChardetWasiApp, embeddedPyChardetWasiRoot+"/app")
		if err != nil {
			pyChardetRuntime.err = fmt.Errorf("pychardet wasm embedded app unavailable: %w", err)
			return
		}

		r := wazero.NewRuntime(ctx)
		if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
			pyChardetRuntime.err = formatPyChardetWasmError("instantiate wasi", err, "")
			_ = r.Close(ctx)
			return
		}

		compiled, err := r.CompileModule(ctx, embeddedPythonWasi)
		if err != nil {
			pyChardetRuntime.err = formatPyChardetWasmError("compile", err, "")
			_ = r.Close(ctx)
			return
		}

		pyChardetRuntime.runtime = r
		pyChardetRuntime.compiled = compiled
		pyChardetRuntime.rootFS = rootFS
		pyChardetRuntime.appFS = appFS
	})
	return pyChardetRuntime.err
}

func (r *pyChardetRuntimeState) fsConfig() wazero.FSConfig {
	return wazero.NewFSConfig().
		WithFSMount(r.rootFS, "/").
		WithFSMount(r.appFS, "/app")
}

func normalizePyChardetEncoding(v string) string {
	v = strings.TrimSpace(v)
	switch strings.ToUpper(v) {
	case "UTF-8-SIG":
		return UTF8WithBOM
	case "IBM866":
		return "cp866"
	case "MACROMAN":
		return "macintosh"
	case "MACCYRILLIC":
		return "x-mac-cyrillic"
	case "SHIFT_JIS":
		return "shift-jis"
	default:
		return strings.ToLower(v)
	}
}

func formatPyChardetWasmError(stage string, err error, stderr string) error {
	if err == nil {
		return nil
	}

	var exitErr *sys.ExitError
	if errors.As(err, &exitErr) {
		err = fmt.Errorf("exit code %d", exitErr.ExitCode())
	}

	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return fmt.Errorf("pychardet wasm %s failed: %w", stage, err)
	}
	return fmt.Errorf("pychardet wasm %s failed: %w: %s", stage, err, stderr)
}
