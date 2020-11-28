package main

import (
	"fmt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	sourceEncoding, _ = parseEncoding("GBK")
	targetEncoding, _ = parseEncoding("UTF8")

	doReplace = false
	doVerbose = false

	files []string
)

func main() {
	parseArgs()

	if len(files) == 0 {
		files = []string{"-"}
		doVerbose = false
	}

	processFiles()
}

const _help = `Examples:
Work with files:
  {exec} [-r] [-s encoding] [-t encoding] [-v] files...
Work with stdio: 
  cat file | {exec}

Arguments:
  -r          Replace/overwrite source file
  -s          Source encoding, default: GBK
  -t          Target encoding, default: UTF8
  -v          Verbose
  -h, --help  Print this help
`

func help() {
	fmt.Println(strings.ReplaceAll(_help, "{exec}", filepath.Base(os.Args[0])))
	os.Exit(0)
}
func exit(err error) {
	log.Print(err)
	os.Exit(-1)
}
func verbose(format string, a ...interface{}) {
	if doVerbose {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}

func parseArgs() {
	var err error
	var args = os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			help()
		case "-s", "-t":
			if i+1 >= len(args) {
				exit(fmt.Errorf("missing encoding [%s] for argument %s", args[i+1], args[i]))
			}
			if args[i] == "-s" {
				sourceEncoding, err = parseEncoding(args[i+1])
			} else {
				targetEncoding, err = parseEncoding(args[i+1])
			}
			if err != nil {
				exit(err)
			}
			i += 1
		case "-r":
			doReplace = true
		case "-v":
			doVerbose = true
		default:
			files = append(files, args[i])
		}
	}
}
func parseEncoding(encoding string) (enc encoding.Encoding, err error) {
	enc, err = htmlindex.Get(encoding)
	if err != nil {
		err = fmt.Errorf("invalid encoding: %s", encoding)
	}
	return
}

func processFiles() {
	var err error

	for _, file := range files {
		verbose("=== %s ===", file)
		err = processFile(file)
		if err != nil {
			log.Print(err)
		}
		verbose("=== %s ===", file)
	}

	return
}
func processFile(file string) (err error) {
	input, output, doDisk := os.Stdin, os.Stdout, file != "-"

	if doDisk {
		{
			var abs string
			verbose("[RESOLVE] %s", file)
			abs, err = filepath.Abs(file)
			if err != nil {
				err = fmt.Errorf("cannot parse file path %s", file)
				return
			}
			verbose("[RESOLVE] %s done: %s", file, abs)

			verbose("[OPEN] %s", abs)
			input, err = os.Open(abs)
			if err == nil {
				defer input.Close()
			} else {
				return
			}
			verbose("[OPEN] %s done", abs)
		}
		{
			verbose("[OPEN] temp file")
			output, err = ioutil.TempFile("", fmt.Sprintf("file_%s.", filepath.Base(input.Name())))
			if err == nil {
				defer output.Close()
			} else {
				return
			}
			verbose("[OPEN] temp file done: %s", output.Name())
		}
	}

	verbose("[TRANSFER] %s => %s", input.Name(), output.Name())
	_, err = io.Copy(
		transform.NewWriter(output, targetEncoding.NewEncoder()),
		transform.NewReader(input, sourceEncoding.NewDecoder()),
	)
	if err != nil {
		err = fmt.Errorf("translate %s => %s failed: %s", input.Name(), output.Name(), err)
		return
	}
	verbose("[TRANSFER] %s => %s done", input.Name(), output.Name())

	if doDisk {
		if doReplace {
			verbose("[REMOVE] %s", input.Name())
			err = os.Remove(input.Name())
			if err != nil {
				err = fmt.Errorf("remove %s failed: %w", input.Name(), err)
				return
			}
			verbose("[REMOVE] %s done", input.Name())

			verbose("[RENAME] %s => %s", output.Name(), input.Name())
			err = os.Rename(output.Name(), input.Name())
			if err != nil {
				err = fmt.Errorf("rename %s => %s failed: %w", output.Name(), input.Name(), err)
				return
			}
			verbose("[RENAME] %s => %s", output.Name(), input.Name())
		} else {
			rename := input.Name() + ".out"
			verbose("[RENAME] %s => %s", output.Name(), rename)
			err = os.Rename(output.Name(), rename)
			if err != nil {
				err = fmt.Errorf("rename %s => %s failed: %w", output.Name(), rename, err)
				return
			}
			verbose("[RENAME] %s => %s", output.Name(), rename)
		}
	}

	return
}
