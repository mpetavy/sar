package main

import (
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	r, _, err := searchAndReplace("abc", "A", "A", false, false, false, false)
	if err != nil || strings.TrimSpace(r) != "abc" {
		t.Fail()
	}

	r, _, err = searchAndReplace("aba", "a", "X", false, false, false, false)
	if err != nil || strings.TrimSpace(r) != "XbX" {
		t.Fail()
	}

	r, _, err = searchAndReplace("abca", "A", "X", true, false, false, false)
	if err != nil || strings.TrimSpace(r) != "XbcX" {
		t.Fail()
	}

	r, _, err = searchAndReplace("abcAbcAB", "ab", "xx", true, true, false, false)
	if err != nil || strings.TrimSpace(r) != "xxcXxcXX" {
		t.Fail()
	}

	r, _, err = searchAndReplace("abcAbcAB", "ab", "xx", true, false, true, false)
	if err != nil || strings.TrimSpace(r) != "XXcXXcXX" {
		t.Fail()
	}

	r, _, err = searchAndReplace("abcAbcAB", "ab", "xx", true, false, false, true)
	if err != nil || strings.TrimSpace(r) != "xxcxxcxx" {
		t.Fail()
	}
}
