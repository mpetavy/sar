package main

import (
	"testing"
)

func TestDefault(t *testing.T) {
	r, err := searchAndReplace("abc", "A", "A", false, false, false, false)
	if err != nil || r != "abc" {
		t.Fail()
	}

	r, err = searchAndReplace("abca", "A", "X", true, false, false, false)
	if err != nil || r != "XbcX" {
		t.Fail()
	}

	r, err = searchAndReplace("abcAbcAB", "ab", "xx", true, true, false, false)
	if err != nil || r != "xxcXxcXX" {
		t.Fail()
	}

	r, err = searchAndReplace("abcAbcAB", "ab", "xx", true, false, true, false)
	if err != nil || r != "XXcXXcXX" {
		t.Fail()
	}

	r, err = searchAndReplace("abcAbcAB", "ab", "xx", true, false, false, true)
	if err != nil || r != "xxcxxcxx" {
		t.Fail()
	}
}
