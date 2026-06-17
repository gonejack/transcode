//go:build charamel_wazero

package chardet

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

const (
	charamelPythonWasmEnv = "CHARDET_CHARAMEL_PYTHON_WASM"
	charamelWasiRootEnv   = "CHARDET_CHARAMEL_WASI_ROOT"
	charamelWasiAppEnv    = "CHARDET_CHARAMEL_WASI_APP"
)

//go:embed charamel_wasi/python.wasm
var embeddedCharamelPythonWasm []byte

//go:embed all:charamel_wasi/app all:charamel_wasi/lib
var embeddedCharamelWasiFS embed.FS

const embeddedCharamelWasiRoot = "charamel_wasi"

type charamelWasiSource struct {
	wasm    []byte
	rootDir string
	rootFS  fs.FS
	appDir  string
	appFS   fs.FS
	homeDir string
}

type charamelRuntimeState struct {
	once     sync.Once
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
	appDir   string
	appFS    fs.FS
	homeDir  string
	rootDir  string
	rootFS   fs.FS
	err      error
}

var charamelWasmRuntime charamelRuntimeState

func DetectEncodingByCharamelWasm(dat []byte) (string, error) {
	if len(dat) == 0 {
		return "", errors.New("charamel wasm: empty input")
	}

	ctx := context.Background()
	if err := initCharamelWasmRuntime(ctx); err != nil {
		return "", err
	}

	var stdout, stderr bytes.Buffer
	mod, err := charamelWasmRuntime.runtime.InstantiateModule(
		ctx,
		charamelWasmRuntime.compiled,
		wazero.NewModuleConfig().
			WithName("").
			WithArgs("python", "-S", "-B", "/app/detect.py").
			WithEnv("PYTHONPATH", "/app").
			WithEnv("PYTHONHOME", charamelWasmRuntime.homeDir).
			WithEnv("PYTHONDONTWRITEBYTECODE", "1").
			WithStdin(bytes.NewReader(dat)).
			WithStdout(&stdout).
			WithStderr(&stderr).
			WithFSConfig(charamelWasmRuntime.fsConfig()).
			WithSysWalltime().
			WithSysNanotime().
			WithSysNanosleep().
			WithRandSource(rand.Reader),
	)
	if mod != nil {
		defer mod.Close(ctx)
	}
	if err != nil {
		return "", formatCharamelWasmError("run", err, stderr.String())
	}

	result := normalizeCharamelWasmEncoding(stdout.String())
	if result == "" {
		return "", formatCharamelWasmError("run", errors.New("empty result"), stderr.String())
	}
	return result, nil
}

func initCharamelWasmRuntime(ctx context.Context) error {
	charamelWasmRuntime.once.Do(func() {
		source, err := charamelWasiRuntime()
		if err != nil {
			charamelWasmRuntime.err = err
			return
		}

		r := wazero.NewRuntime(ctx)
		if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
			charamelWasmRuntime.err = formatCharamelWasmError("instantiate wasi", err, "")
			_ = r.Close(ctx)
			return
		}

		compiled, err := r.CompileModule(ctx, source.wasm)
		if err != nil {
			charamelWasmRuntime.err = formatCharamelWasmError("compile", err, "")
			_ = r.Close(ctx)
			return
		}

		charamelWasmRuntime.runtime = r
		charamelWasmRuntime.compiled = compiled
		charamelWasmRuntime.rootDir = source.rootDir
		charamelWasmRuntime.rootFS = source.rootFS
		charamelWasmRuntime.appDir = source.appDir
		charamelWasmRuntime.appFS = source.appFS
		charamelWasmRuntime.homeDir = source.homeDir
	})
	return charamelWasmRuntime.err
}

func (r *charamelRuntimeState) fsConfig() wazero.FSConfig {
	config := wazero.NewFSConfig()
	if r.rootFS != nil {
		config = config.WithFSMount(r.rootFS, "/")
	} else {
		config = config.WithReadOnlyDirMount(r.rootDir, "/")
	}
	if r.appFS != nil {
		config = config.WithFSMount(r.appFS, "/app")
	} else {
		config = config.WithReadOnlyDirMount(r.appDir, "/app")
	}
	return config
}

func charamelWasiRuntime() (charamelWasiSource, error) {
	var source charamelWasiSource
	pythonWasm := os.Getenv(charamelPythonWasmEnv)
	if pythonWasm == "" {
		source.wasm = embeddedCharamelPythonWasm
	} else {
		wasm, err := os.ReadFile(pythonWasm)
		if err != nil {
			return source, fmt.Errorf("charamel wasm read %s failed: %w", pythonWasm, err)
		}
		source.wasm = wasm
	}

	rootDir := os.Getenv(charamelWasiRootEnv)
	if rootDir != "" {
		if _, statErr := os.Stat(rootDir); statErr != nil {
			return source, fmt.Errorf("charamel wasm missing %s path %s: %w", charamelWasiRootEnv, rootDir, statErr)
		}
		if _, statErr := os.Stat(path.Join(rootDir, "lib")); statErr != nil {
			return source, fmt.Errorf("charamel wasm missing %s lib directory under %s: %w", charamelWasiRootEnv, rootDir, statErr)
		}
		source.rootDir = rootDir
		source.homeDir = "/"
	} else {
		rootFS, err := fs.Sub(embeddedCharamelWasiFS, embeddedCharamelWasiRoot)
		if err != nil {
			return source, fmt.Errorf("charamel wasm embedded root unavailable: %w", err)
		}
		source.rootFS = rootFS
		source.homeDir = "/"
	}

	appDir := os.Getenv(charamelWasiAppEnv)
	if appDir != "" {
		if _, statErr := os.Stat(appDir); statErr != nil {
			return source, fmt.Errorf("charamel wasm missing %s path %s: %w", charamelWasiAppEnv, appDir, statErr)
		}
		source.appDir = appDir
	} else {
		appFS, err := fs.Sub(embeddedCharamelWasiFS, path.Join(embeddedCharamelWasiRoot, "app"))
		if err != nil {
			return source, fmt.Errorf("charamel wasm embedded app unavailable: %w", err)
		}
		source.appFS = appFS
	}
	return source, nil
}

func normalizeCharamelWasmEncoding(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	switch v {
	case "utf_8":
		return "utf-8"
	case "utf_8_sig":
		return "utf-8-bom"
	case "utf_16":
		return "utf-16"
	case "utf_16_be":
		return "utf-16be"
	case "utf_16_le":
		return "utf-16le"
	case "utf_32":
		return "utf-32"
	case "utf_32_be":
		return "utf-32be"
	case "utf_32_le":
		return "utf-32le"
	case "big5hkscs":
		return "big5-hkscs"
	case "euc_jp":
		return "euc-jp"
	case "euc_kr":
		return "euc-kr"
	case "shift_jis":
		return "shift-jis"
	case "iso2022_jp":
		return "iso-2022-jp"
	case "iso2022_kr":
		return "iso-2022-kr"
	case "koi8_r":
		return "koi8-r"
	case "koi8_u":
		return "koi8-u"
	case "tis_620":
		return "tis-620"
	case "latin_1":
		return "iso-8859-1"
	default:
		return v
	}
}

func formatCharamelWasmError(stage string, err error, stderr string) error {
	if err == nil {
		return nil
	}

	var exitErr *sys.ExitError
	if errors.As(err, &exitErr) {
		err = fmt.Errorf("exit code %d", exitErr.ExitCode())
	}

	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return fmt.Errorf("charamel wasm %s failed: %w", stage, err)
	}
	return fmt.Errorf("charamel wasm %s failed: %w: %s", stage, err, stderr)
}
