package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/mpetavy/common"
)

var (
	searchStr         *string
	replaceStr        *string
	filemask          *string
	ignoreError       *bool
	ignoreCase        *bool
	ignoreHidden      *bool
	negative          *bool
	replaceCase       *bool
	recursive         *bool
	replaceUpper      *bool
	replaceLower      *bool
	dryrun            *bool
	onlyListFilenames *bool
	plain             *bool
)

func init() {
	common.Init(false, "1.0.8", "", "", "2018", "Simple search and replace", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, nil, run, 0)

	searchStr = flag.String("s", "", "search text")
	replaceStr = flag.String("t", "", "replace text")
	filemask = flag.String("f", "*", "input file or STDIN")
	negative = flag.Bool("n", false, "negative search")
	ignoreError = flag.Bool("e", true, "ignore error")
	ignoreCase = flag.Bool("i", false, "ignore case")
	ignoreHidden = flag.Bool("x", true, "ignore hidden directories")
	recursive = flag.Bool("r", false, "recursive directory search")
	replaceUpper = flag.Bool("tu", false, "replace to replaceUpper")
	replaceLower = flag.Bool("tl", false, "replace to replaceLower")
	replaceCase = flag.Bool("tc", false, "replace case sensitive like found text")
	dryrun = flag.Bool("d", false, "dry run")
	onlyListFilenames = flag.Bool("l", false, "only list files")
	plain = flag.Bool("p", false, "plain output")
}

func searchAndReplace(input string, searchStr string, replaceStr string, ignoreCase bool, replaceCase bool, replaceUpper bool, replaceLower bool) (string, []string, error) {
	lines := []string{}
	output := ""
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(common.ScanLinesWithLF)

	regexValue := searchStr
	if ignoreCase {
		regexValue = fmt.Sprintf("(?i)%s", searchStr)
	}

	regex := regexp.MustCompile(regexValue)

	c := 0
	lc := 0
	for scanner.Scan() {
		lc++

		line := scanner.Text()
		oldLine := line

		diffAdd := 0
		indices := regex.FindAllIndex([]byte(line), -1)

		for i, index := range indices {
			c++

			foundLen := index[1] - index[0]

			p := indices[i][0] + diffAdd

			if *plain {
				lines = append(lines, line[index[0]:index[1]]+"\n")
			} else {
				lines = append(lines, fmt.Sprintf("line %5d: %s", lc, line))
			}

			if replaceStr == "" {
				break
			}

			txt := line[p : p+foundLen]
			isLetter := false
			firstUpper := false
			secondUpper := false

			if replaceCase {
				r, err := common.Rune(txt, 0)
				if err != nil {
					return "", lines, err
				}
				firstUpper = unicode.IsUpper(r)
				isLetter = unicode.IsLetter(r)

				if len(txt) > 1 {
					r, err := common.Rune(txt, 1)
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

			line = line[:p] + replaceStr + line[p+foundLen:]

			diffAdd += len(replaceStr) - foundLen
		}

		output = output + line

		if oldLine != line {
			if *plain {
				lines = append(lines, line)
			} else {
				lines = append(lines, fmt.Sprintf("line %5d: %s", lc, line))
			}
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

func processFile(filename string, f os.FileInfo) error {
	if f.IsDir() {
		return nil
	}

	common.DebugFunc(filename)
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	input := string(b)

	output, lines, err := searchAndReplace(input, *searchStr, *replaceStr, *ignoreCase, *replaceCase, *replaceUpper, *replaceLower)
	if err != nil {
		return err
	}

	if !*plain && len(lines) > 0 != *negative {
		fmt.Printf("%s\n", filename)
	}

	if len(lines) > 0 && !*negative {
		if !*onlyListFilenames {
			for _, l := range lines {
				fmt.Printf("%s", l)
			}
		}
	}

	if !*dryrun && output != input && *replaceStr != "" {
		err = common.FileBackup(filename)
		if err != nil {
			return err
		}

		err = os.WriteFile(filename, []byte(output), common.DefaultFileMode)
		if err != nil {
			return err
		}
	}

	return err
}

func run() error {
	if *filemask == "" {
		err := processStream(os.Stdin, os.Stdout)
		if err != nil {
			return err
		}

		return nil
	}

	fw, err := common.NewFilewalker(*filemask, *recursive, *ignoreError, processFile)
	if err != nil {
		return err
	}

	return fw.Run()
}

func main() {
	defer common.Done()

	common.Run([]string{"s"})
}
