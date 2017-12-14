package util

import (
	"fmt"
	"os"

	flag "github.com/puellanivis/breton/lib/gnuflag"
)

// CleanLineStart is a string that may be printed which will reset to column 0, and then set ANSI formatting to default.
const CleanLineStart = "\r\033[K"

// Status is a convenience shortcut for fmt.Fprint(os.Stderr, ...)
func Status(values ...interface{}) {
	fmt.Fprint(os.Stderr, values...)
}

// Statusln is a convenience shortcut for fmt.Fprintln(os.Stderr, ...)
func Statusln(values ...interface{}) {
	fmt.Fprintln(os.Stderr, values...)
}

// Statusf is a convenience shortcut for fmt.Fprintf(os.Stderr, format, ...)
func Statusf(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, format, values...)
}

// Fatal is a convenience shortcut for fmt.Fprint(os.Stderr, ...) followed by this library's Exit(1).
func Fatal(values ...interface{}) {
	Status(values...)
	Exit(1)
}

// Fatalln is a convenience shortcut for fmt.Fprintln(os.Stderr, ...) followed by this library's Exit(1).
func Fatalln(values ...interface{}) {
	Statusln(values...)
	Exit(1)
}

// Fatalf is a convenience shortcut for fmt.Fprintf(os.Stderr, format, ...) followed by this library's Exit(1).
func Fatalf(format string, values ...interface{}) {
	Statusf(format, values...)
	Exit(1)
}
