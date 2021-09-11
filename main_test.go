package main

import (
	"fmt"
	"testing"
)

func TestDefault(t *testing.T) {
	var tests = []struct {
		input        string
		searchStr    string
		replaceStr   string
		ignoreCase   bool
		replaceCase  bool
		replaceUpper bool
		replaceLower bool
		result       string
	}{
		{
			input:        "abc",
			searchStr:    "A",
			replaceStr:   "X",
			ignoreCase:   false,
			replaceCase:  false,
			replaceUpper: false,
			replaceLower: false,
			result:       "abc",
		},
		{
			input:        "aba",
			searchStr:    "a",
			replaceStr:   "X",
			ignoreCase:   false,
			replaceCase:  false,
			replaceUpper: false,
			replaceLower: false,
			result:       "XbX",
		},
		{
			input:        "abca",
			searchStr:    "A",
			replaceStr:   "X",
			ignoreCase:   true,
			replaceCase:  false,
			replaceUpper: false,
			replaceLower: false,
			result:       "XbcX",
		},
		{
			input:        "abcAbcAB",
			searchStr:    "ab",
			replaceStr:   "xx",
			ignoreCase:   true,
			replaceCase:  true,
			replaceUpper: false,
			replaceLower: false,
			result:       "xxcXxcXX",
		},
		{
			input:        "abcAbcAB",
			searchStr:    "ab",
			replaceStr:   "xx",
			ignoreCase:   true,
			replaceCase:  false,
			replaceUpper: true,
			replaceLower: false,
			result:       "XXcXXcXX",
		},
		{
			input:        "abcAbcAB",
			searchStr:    "ab",
			replaceStr:   "xx",
			ignoreCase:   true,
			replaceCase:  false,
			replaceUpper: false,
			replaceLower: true,
			result:       "xxcxxcxx",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test %d: %+v", i, test), func(t *testing.T) {
			r, _, err := searchAndReplace(test.input, test.searchStr, test.replaceStr, test.ignoreCase, test.replaceCase, test.replaceUpper, test.replaceLower)
			if err != nil {
				t.Error(err)
			}

			if r != test.result {
				t.Errorf("got %q, want %q", r, test.result)
			}
		})
	}
}
