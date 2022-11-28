// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/ (With no dependency injection, the code had to be modified to integrate with github.com/puellanivis/breton/lib/gnuflag)
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package glog implements logging analogous to the Google-internal C++ INFO/ERROR/V setup.
// It provides functions Info, Warning, Error, Fatal, plus formatting variants such as
// Infof. It also provides V-style logging controlled by the --verbosity and --vmodule=file=2 flags.
//
// Basic examples:
//
//	glog.Info("Prepare to repel boarders")
//
//	glog.Fatalf("Initialization failed: %s", err)
//
// See the documentation for the V function for an explanation of these examples:
//
//	if glog.V(2) {
//		log.Info("Starting transaction...")
//	}
//
//	glog.V(2).Infoln("Processed", nItems, "elements")
//
// Log output is buffered and written periodically using Flush. Programs
// should call Flush before exiting to guarantee all log output is written.
//
// By default, all log statements write to files in a temporary directory.
// This package provides several flags that modify this behavior.
// As a result, flag.Parse must be called before any logging is done.
//
//	--logtostderr=false
//		Logs are written to standard error instead of to files.
//	--alsologtostderr=false
//		Logs are written to standard error as well as to files.
//	--stderrthreshold=ERROR
//		Log events at or above this severity are logged to standard
//		error as well as to files.
//	--log_dir=""
//		Log files will be written to this directory instead of the
//		default temporary directory.
//
//	Other flags provide aids to debugging.
//
//	--log_backtrace_at=""
//		When set to a file and line number holding a logging statement,
//		such as
//			--log_backtrace_at=gopherflakes.go:234
//		a stack trace will be written to the Info log whenever execution
//		hits that statement. (Unlike with --vmodule, the ".go" must be
//		present.)
//	--verbosity=0
//		Enable V-leveled logging at the specified level.
//	--vmodule=""
//		The syntax of the argument is a comma-separated list of pattern=N,
//		where pattern is a literal file name (minus the ".go" suffix) or
//		"glob" pattern and N is a V level. For instance,
//			--vmodule=gopher*=3
//		sets the V level to 3 in all Go files whose names begin "gopher".
package glog

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	stdLog "log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	flag "github.com/puellanivis/breton/lib/gnuflag"
)

// flushSyncWriter is the interface satisfied by logging destinations.
type flushSyncWriter interface {
	Flush() error
	Sync() error
	io.Writer
}

const flushInterval = 30 * time.Second

func init() {
	go func() {
		for range time.Tick(flushInterval) {
			Flush()
		}
	}()
}

// loggingT collects all the global state of the logging setup.
type loggingT struct {
	// This is used to synchronize logging.
	sync.Mutex

	flagT

	// file holds writer for each of the log types.
	file [numSeverity]flushSyncWriter
}

var logging loggingT

// Flush flushes all pending log I/O.
func Flush() {
	logging.Lock()
	defer logging.Unlock()

	logging.flushAll()
}

// flushAll flushes all the logs and attempts to "sync" their data to disk.
// One is expected to be holding the loggingT.Mutex
func (l *loggingT) flushAll() {
	// Flush from fatal down, in case there's trouble flushing.
	for s := fatalLog; s >= infoLog; s-- {
		if file := l.file[s]; file != nil {
			_ = file.Flush()
			_ = file.Sync()
		}
	}
}

var timeNow = time.Now // Stubbed out for testing.

/*
header formats a log header as defined by the C++ implementation.
It returns a buffer containing the formatted header and the user's file and line number.
The depth specifies how many stack frames above lives the source line to be identified in the log message.

Log lines have this form:

	Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...

where the fields are defined as follows:

	L                A single character, representing the log level (eg 'I' for INFO)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	file             The file name
	line             The line number
	msg              The user-supplied message
*/
func (l *loggingT) header(s severity, depth int) (*bytes.Buffer, string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	}

	if slash := strings.LastIndexByte(file, '/'); slash >= 0 {
		file = file[slash+1:]
	}

	return l.formatHeader(s, file, line), file, line
}

// formatHeader formats a log header using the provided file name and line number.
func (l *loggingT) formatHeader(s severity, file string, line int) *bytes.Buffer {
	now := timeNow()

	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}

	if s > fatalLog {
		s = infoLog // for safety.
	}
	buf := getBuffer()

	// Thinks are simple enough, Fprintf uses a lot of reflection that is unnecessary,
	// but bytes.Buffer already uses a small bootstrap array,
	// and strconv.Itoa has some amazingly clever speed up to handle small values.
	_, month, day := now.Date()
	hour, minute, second := now.Clock()
	ns := now.Nanosecond() / 1000

	// Lmmdd hh:mm:ss.uuuuuu threadid file:line]
	buf.WriteByte(severityChars[s])
	writeTwo(buf, int(month))
	writeTwo(buf, day)
	buf.WriteByte(' ')

	writeTwo(buf, hour)
	buf.WriteByte(':')
	writeTwo(buf, minute)
	buf.WriteByte(':')
	writeTwo(buf, second)
	buf.WriteByte('.')

	nano := strconv.Itoa(ns)
	if len(nano) < 6 {
		buf.WriteString("000000"[len(nano):])
	}
	buf.WriteString(nano)
	buf.WriteByte(' ')

	buf.Write(tid)
	buf.WriteByte(' ')

	buf.WriteString(file)
	buf.WriteByte(':')

	buf.WriteString(strconv.Itoa(line))
	buf.WriteByte(']')
	buf.WriteByte(' ')

	return buf
}

func (l *loggingT) println(s severity, args ...interface{}) {
	buf, file, line := l.header(s, 0)
	fmt.Fprintln(buf, args...)
	l.output(s, buf, file, line, false)
}

func (l *loggingT) printDepth(s severity, depth int, args ...interface{}) {
	buf, file, line := l.header(s, depth)

	fmt.Fprint(buf, args...)
	ensureNL(buf)

	l.output(s, buf, file, line, false)
}

func (l *loggingT) print(s severity, args ...interface{}) {
	l.printDepth(s, 1, args...)
}

func (l *loggingT) printf(s severity, format string, args ...interface{}) {
	buf, file, line := l.header(s, 0)

	fmt.Fprintf(buf, format, args...)
	ensureNL(buf)

	l.output(s, buf, file, line, false)
}

// output writes the data to the log files and releases the buffer.
func (l *loggingT) output(s severity, buf *bytes.Buffer, file string, line int, alsoToStderr bool) {
	if l.TraceLocation.isSet() {
		if l.TraceLocation.match(file, line) {
			buf.Write(stacks(false))
		}
	}
	data := buf.Bytes()

	l.Lock()

	if !flag.Parsed() {
		os.Stderr.Write([]byte("ERROR: logging before flag.Parse: "))
		os.Stderr.Write(data)

	} else if l.ToStderr {
		os.Stderr.Write(data)

	} else {
		if alsoToStderr || l.AlsoToStderr || s >= l.StderrThreshold.get() {
			os.Stderr.Write(data)
		}

		if l.file[s] == nil {
			if err := l.createFiles(s); err != nil {
				os.Stderr.Write(data) // Make sure the message appears somewhere.
				l.exit(err)
			}
		}

		switch s {
		case fatalLog:
			_, _ = l.file[fatalLog].Write(data)
			fallthrough
		case errorLog:
			_, _ = l.file[errorLog].Write(data)
			fallthrough
		case warningLog:
			_, _ = l.file[warningLog].Write(data)
			fallthrough
		case infoLog:
			_, _ = l.file[infoLog].Write(data)
		}
	}

	if s == fatalLog {
		// If we got here via Exit rather than Fatal, print no stacks.
		if atomic.LoadUint32(&fatalNoStacks) > 0 {
			l.Unlock()

			timeoutFlush(10 * time.Second)
			os.Exit(1)
		}

		// Dump all goroutine stacks before exiting.
		// First, make sure we see the trace for the current goroutine on standard error.
		// If --logtostderr has been specified, the loop below will do that anyway
		// as the first stack in the full dump.
		if !l.ToStderr {
			os.Stderr.Write(stacks(false))
		}

		// Write the stack trace for all goroutines to the files.
		trace := stacks(true)

		for log := fatalLog; log >= infoLog; log-- {
			if f := l.file[log]; f != nil { // Can be nil if -logtostderr is set.
				_, _ = f.Write(trace)
			}
		}

		l.Unlock()

		timeoutFlush(10 * time.Second)
		os.Exit(255) // C++ uses -1, which is silly because it's anded with 255 anyway.
	}

	l.Unlock()

	putBuffer(buf)
	if stats := severityStats[s]; stats != nil {
		stats.add(int64(len(data)))
	}
}

// timeoutFlush calls Flush and returns when it completes or after timeout
// elapses, whichever happens first.  This is needed because the hooks invoked
// by Flush may deadlock when glog.Fatal is called from a hook that holds
// a lock.
func timeoutFlush(timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		defer close(done)

		Flush() // calls logging.lockAndFlushAll()
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		fmt.Fprintln(os.Stderr, "glog: Flush took longer than", timeout)
	}
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

// exit is called if there is trouble creating or writing log files.
// It flushes the logs and exits the program; there's no point in hanging around.
// logging Mutex is held.
func (l *loggingT) exit(err error) {
	fmt.Fprintln(os.Stderr, "log: exiting because of error: ", err)
	Flush()
	os.Exit(2)
}

// syncBuffer joins a bufio.Writer to its underlying file, providing access to the
// file's Sync method and providing a wrapper for the Write method that provides log
// file rotation. There are conflicting methods, so the file cannot be embedded.
// logging Mutex is held for all its methods.
type syncBuffer struct {
	logger *loggingT
	sev    severity

	*bufio.Writer
	file   *os.File
	nbytes uint64 // The number of bytes written to this file
}

func (sb *syncBuffer) Sync() error {
	return sb.file.Sync()
}

func (sb *syncBuffer) Write(b []byte) (n int, err error) {
	if sb.nbytes+uint64(len(b)) >= MaxSize {
		if err := sb.rotateFile(time.Now()); err != nil {
			sb.logger.exit(err)
		}
	}

	n, err = sb.Writer.Write(b)
	sb.nbytes += uint64(n)
	if err != nil {
		sb.logger.exit(err)
	}

	return n, err
}

// bufferSize sizes the buffer associated with each log file. It's large
// so that log records can accumulate without the logging thread blocking
// on disk I/O. The flushDaemon will block instead.
const bufferSize = 256 * 1024

// rotateFile closes the syncBuffer's file and starts a new one.
func (sb *syncBuffer) rotateFile(now time.Time) error {
	if sb.file != nil {
		if err := sb.Flush(); err != nil {
			return err
		}

		if err := sb.file.Close(); err != nil {
			return err
		}
	}

	var err error
	sb.file, _, err = create(severityNames[sb.sev], now)
	sb.nbytes = 0
	if err != nil {
		return err
	}

	sb.Writer = bufio.NewWriterSize(sb.file, bufferSize)

	// Write header.
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "Log file created at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(buf, "Running on machine: %s\n", host)
	fmt.Fprintf(buf, "Binary: Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(buf, "Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg\n")

	n, err := buf.WriteTo(sb.file)
	sb.nbytes += uint64(n)
	return err
}

// createFiles creates all the log files for severity from sev down to infoLog.
// logging mutex is held.
func (l *loggingT) createFiles(sev severity) error {
	now := time.Now()

	// Files are created in decreasing severity order, so as soon as we find one
	// has already been created, we can stop.
	for s := sev; s >= infoLog && l.file[s] == nil; s-- {
		sb := &syncBuffer{
			logger: l,
			sev:    s,
		}

		if err := sb.rotateFile(now); err != nil {
			return err
		}

		l.file[s] = sb
	}

	return nil
}

// CopyStandardLogTo arranges for messages written to the Go "log" package's
// default logs to also appear in the Google logs for the named and lower
// severities.  Subsequent changes to the standard log's default output location
// or format may break this behavior.
//
// Valid names are "INFO", "WARNING", "ERROR", and "FATAL".  If the name is not
// recognized, CopyStandardLogTo panics.
func CopyStandardLogTo(name string) {
	sev, ok := severityByName[name]
	if !ok {
		panic(fmt.Sprintf("log.CopyStandardLogTo(%q): unrecognized severity name", name))
	}

	// Set a log format that captures the user's file and line:
	//   d.go:23: message
	stdLog.SetFlags(stdLog.Lshortfile)
	stdLog.SetOutput(logBridge(sev))
}

// logBridge provides the Write method that enables CopyStandardLogTo to connect
// Go's standard logs to the logs provided by this package.
type logBridge severity

// Write parses the standard logging line and passes its components to the
// logger for severity(lb).
func (lb logBridge) Write(b []byte) (n int, err error) {
	file, line := "???", 1
	var text string

	// Split "d.go:23: message" into "d.go", "23", and "message".
	if parts := bytes.SplitN(b, []byte{':'}, 3); len(parts) != 3 || len(parts[0]) < 1 || len(parts[2]) < 1 {
		text = fmt.Sprintf("bad log format: %s", b)

	} else {
		file = string(parts[0])
		text = string(parts[2][1:]) // skip leading space

		line, err = strconv.Atoi(string(parts[1]))
		if err != nil {
			text = fmt.Sprintf("bad line number: %s", b)
			line = 1
		}
	}

	buf := logging.formatHeader(severity(lb), file, line)

	buf.WriteString(text)
	ensureNL(buf)

	// l.output with alsoToStderr=true, so standard log messages
	// always appear on standard error.
	logging.output(severity(lb), buf, file, line, true)

	return len(b), nil
}
