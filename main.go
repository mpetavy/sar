package main

import (
	"bufio"
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
	searchStr         *string
	replaceStr        *string
	filemask          *string
	ignoreCase        *bool
	replaceCase       *bool
	recursive         *bool
	replaceUpper      *bool
	replaceLower      *bool
	backup            *bool
	simulate          *bool
	onlyListFilenames *bool
	rootPath          string
)

func init() {
	searchStr = flag.String("s", "", "search text")
	replaceStr = flag.String("t", "", "replace text")
	filemask = flag.String("f", "", "input file or STDIN")
	ignoreCase = flag.Bool("i", false, "ignore case")
	recursive = flag.Bool("r", false, "recursive directory search")
	replaceUpper = flag.Bool("tu", false, "replace to replaceUpper")
	replaceLower = flag.Bool("tl", false, "replace to replaceLower")
	replaceCase = flag.Bool("tc", false, "replace case sensitive like found text")
	backup = flag.Bool("b", true, "create backup files")
	simulate = flag.Bool("d", false, "dry run")
	onlyListFilenames = flag.Bool("l", false, "only list files")
}

func prepare() error {
	common.NoBanner = *filemask == ""

	return nil
}

func searchAndReplace(input string, searchStr string, replaceStr string, ignoreCase bool, replaceCase bool, replaceUpper bool, replaceLower bool) (string, []string, error) {
	lines := []string{}
	output := ""
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(common.ScanLinesWithLF)

	c := 0
	for scanner.Scan() {
		line := scanner.Text()
		oldLine := line

		c++

		nextP := 0
		p := 0

		for {
			if ignoreCase {
				p = strings.Index(strings.ToUpper(line[nextP:]), strings.ToUpper(searchStr))
			} else {
				p = strings.Index(line[nextP:], searchStr)
			}

			if p == -1 {
				break
			}

			p += nextP
			if replaceStr != "" {
				nextP = p + len(replaceStr)
			} else {
				nextP = p + len(searchStr)
			}

			lines = append(lines, fmt.Sprintf("%5d: %s", c, line))

			if replaceStr == "" {
				break
			}

			txt := line[p : p+len(searchStr)]
			isLetter := false
			firstUpper := false
			secondUpper := false

			if replaceCase {
				r, err := common.GetRune(txt, 0)
				if err != nil {
					return "", lines, err
				}
				firstUpper = unicode.IsUpper(r)
				isLetter = unicode.IsLetter(r)

				if len(txt) > 1 {
					r, err := common.GetRune(txt, 1)
					if err != nil {
						return "", lines, err
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

			if replaceUpper {
				replaceStr = strings.ToUpper(replaceStr)
			}

			if replaceLower {
				replaceStr = strings.ToLower(replaceStr)
			}

			line = line[:p] + replaceStr + line[p+len(searchStr):]
		}

		output = output + line

		if oldLine != line {
			lines = append(lines, fmt.Sprintf("%5d: %s", c, line))
		}
	}

	return output, lines, nil
}

func processStream(input io.Reader, output io.Writer) error {
	b := bytes.Buffer{}

	_, err := io.Copy(&b, input)
	if err != nil {
		return err
	}

	str := string(b.Bytes())
	str, _, err = searchAndReplace(str, *searchStr, *replaceStr, *ignoreCase, *replaceCase, *replaceUpper, *replaceLower)
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

	input := string(b)

	output, lines, err := searchAndReplace(input, *searchStr, *replaceStr, *ignoreCase, *replaceCase, *replaceUpper, *replaceLower)
	if err != nil {
		return err
	}

	if len(lines) > 0 {
		fmt.Printf("%s\n", filename)
		if !*onlyListFilenames {
			for _, l := range lines {
				fmt.Printf("%s", l)
			}
		}
	}

	if !*simulate && output != input {
		if *replaceStr != "" && *backup {
			err = common.FileBackup(filename, 1)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(filename, []byte(output), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func walkfunc(path string, _ os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	b, err := common.IsFile(path)
	if err != nil {
		return err
	}

	if b {
		b = *filemask == ""

		if !b {
			b, err = common.EqualWildcards(filepath.Base(path), *filemask)
			if err != nil {
				return err
			}
		}

		if !b {
			return nil
		}

		return processFile(path)
	}

	if *recursive || path == rootPath {
		return nil
	}

	return filepath.SkipDir
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

	*filemask = common.CleanPath(*filemask)

	if common.ContainsWildcard(*filemask) {
		rootPath = filepath.Dir(*filemask)
		*filemask = filepath.Base(*filemask)
	} else {
		b, err := common.FileExists(*filemask)
		if err != nil {
			return err
		}

		if b {
			b, err := common.IsDirectory(*filemask)
			if err != nil {
				return err
			}

			if b {
				rootPath = *filemask
				*filemask = ""
			} else {
				rootPath = filepath.Dir(*filemask)
				*filemask = filepath.Base(*filemask)
			}
		}
	}

	return walk(rootPath)
}

func main() {
	defer common.Cleanup()

	common.New(&common.App{"sar", "1.0.6", "2018", "Simple search and replace", "mpetavy", common.APACHE, "https://github.com/mpetavy/sar", false, prepare, nil, nil, run, time.Duration(0)}, []string{"s"})
	common.Run()
}
