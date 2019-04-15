package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/mpetavy/common"
)

var (
	searchStr    *string
	replaceStr   *string
	filemask     *string
	ignoreCase   *bool
	replaceCase  *bool
	recursive    *bool
	replaceUpper *bool
	replaceLower *bool
	backup       *bool

	rootPath string
)

func init() {
	searchStr = flag.String("s", "", "search text")
	replaceStr = flag.String("r", "", "replace text")
	filemask = flag.String("f", "", "input file or STDIN")
	ignoreCase = flag.Bool("i", false, "ignore case")
	recursive = flag.Bool("R", false, "recursive directory search")
	replaceUpper = flag.Bool("tu", false, "replace to replaceUpper")
	replaceLower = flag.Bool("tl", false, "replace to replaceLower")
	replaceCase = flag.Bool("tc", false, "replace case sensitive")
	backup = flag.Bool("b", true, "create backup files")
}

func prepare() error {
	common.NoBanner = *filemask == ""

	return nil
}

func searchAndReplace(str string, searchStr string, replaceStr string, ignoreCase bool, replaceCase bool, upper bool, lower bool) (string, error) {
	_str := str

	if ignoreCase {
		_str = strings.ToUpper(_str)
		searchStr = strings.ToUpper(searchStr)
	}

	for {
		p := strings.Index(_str, searchStr)

		if p == -1 {
			break
		}

		txt := str[p : p+len(searchStr)]
		isLetter := false
		firstUpper := false
		secondUpper := false

		if replaceCase {
			r, err := common.GetRune(txt, 0)
			if err != nil {
				return "", err
			}
			firstUpper = unicode.IsUpper(r)
			isLetter = unicode.IsLetter(r)

			if len(txt) > 1 {
				r, err := common.GetRune(txt, 1)
				if err != nil {
					return "", err
				}
				secondUpper = unicode.IsUpper(r)
			}

			if isLetter {
				if firstUpper {
					if secondUpper {
						replaceStr = strings.ToUpper(replaceStr)
					} else {
						replaceStr = common.Capitalize(replaceStr)
					}
				} else {
					replaceStr = strings.ToLower(replaceStr)
				}
			}
		}

		if upper {
			replaceStr = strings.ToUpper(replaceStr)
		}

		if lower {
			replaceStr = strings.ToLower(replaceStr)
		}

		_str = _str[:p] + replaceStr + _str[p+len(searchStr):]
		str = str[:p] + replaceStr + str[p+len(searchStr):]
	}

	return str, nil
}

func processStream(input io.Reader, output io.Writer) error {
	b := bytes.Buffer{}

	_, err := io.Copy(&b, input)
	if err != nil {
		return err
	}

	str := string(b.Bytes())
	str, err = searchAndReplace(str, *searchStr, *replaceStr, *ignoreCase, *replaceCase, *replaceUpper, *replaceLower)
	if err != nil {
		return err
	}

	_, err = output.Write([]byte(str))
	if err != nil {
		return err
	}

	return nil
}

func processFile(filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	str := string(b)
	str, err = searchAndReplace(str, *searchStr, *replaceStr, *ignoreCase, *replaceCase, *replaceUpper, *replaceLower)
	if err != nil {
		return err
	}

	if str != string(b) {
		fmt.Printf("%s\n", filename)

		if *replaceStr != "" {
			err = common.FileBackup(filename, 1)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(filename, []byte(str), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func walkfunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if common.IsFile(path) {
		var b bool

		b, err = common.EqualWildcards(filepath.Base(path), *filemask)
		if err != nil {
			return err
		}

		if !b {
			return nil
		}
		return processFile(path)
	} else {
		if *recursive || path == rootPath {
			return nil
		}

		return filepath.SkipDir
	}
}

func walk(path string) error {
	err := filepath.Walk(path, walkfunc)
	if err != nil {
		return err
	}

	return nil
}

func run() error {
	if *filemask == "" {
		err := processStream(os.Stdin, os.Stdout)
		if err != nil {
			return err
		}

		return nil
	}

	var err error

	*filemask = common.CleanPath(*filemask)

	if common.ContainsWildcard(*filemask) {
		rootPath = filepath.Dir(*filemask)

		if rootPath == "." {
			rootPath, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		*filemask = filepath.Base(*filemask)

		err = walk(rootPath)
		if err != nil {
			return err
		}

		return nil
	}

	b, err := common.FileExists(*filemask)
	if err != nil {
		return err
	}

	if b {
		return processFile(*filemask)
	}

	return fmt.Errorf("cannot process: %s", *filemask)
}

func main() {
	defer common.Cleanup()

	common.New(&common.App{"sar", "1.0.0", "2018", "Simple search and replace", "mpetavy", common.APACHE, "https://github.com/mpetavy/sar", false, prepare, nil, nil, run, time.Duration(0)}, []string{"s"})
	common.Run()
}
