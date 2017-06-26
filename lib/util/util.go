package util

import (
	"lib/sort"
)

// Nope does nothing, but "uses" a variable
func Nope(v ...interface{}) {
	return
}

// CopyBytes returns a wholey new copy of the []byte passed in
func CopyBytes(b []byte) []byte {
	buf := make([]byte, len(b))
	copy(buf, b)

	return buf
}

// UniqueStrings returns a sorted list of values where duplicates have been removed.
func UniqueStrings(values []string) []string {
	m := make(map[string]bool)

	for _, v := range values {
		m[v] = true
	}

	var r []string

	for k := range m {
		r = append(r, k)
	}

	sort.Strings(r)

	return r
}

// StripNL strips a newline from the byte slice given.
func StripNL(line []byte) []byte {
	if len(line) < 1 {
		return line
	}

	i := len(line) - 1
	if line[i] != '\n' {
		return line
	}

	return line[:i]
}

// Forever blocks and never returns.
func Forever() {
	var ch chan struct{}

	<-ch
}