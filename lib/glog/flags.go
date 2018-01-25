package glog

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	flag "github.com/puellanivis/breton/lib/gnuflag"
)

// severity identifies the sort of log: info, warning, etc.
// It also implements the flag.Value interface.
// The --stderrthreshold flag is of type severity and should be modified only through the flag.Value interface.
// The values match the corresponding constants in C++.
type severity int32 // sync/atomic int32

// These constants identify the log levels in order of increasing severity.
// A message written to a high-severity log file is also written to each
// lower-severity log file.
const (
	infoLog severity = iota
	warningLog
	errorLog
	fatalLog
	numSeverity = 4
)

const severityChars = "IWEF"

var severityNames = []string{
	infoLog:    "INFO",
	warningLog: "WARNING",
	errorLog:   "ERROR",
	fatalLog:   "FATAL",
}

var severityByName = map[string]severity{
	"INFO":    infoLog,
	"WARNING": warningLog,
	"ERROR":   errorLog,
	"FATAL":   fatalLog,
}

// get returns the value of the severity.
func (s *severity) get() severity {
	return severity(atomic.LoadInt32((*int32)(s)))
}

// set sets the value of the severity.
func (s *severity) set(val severity) {
	atomic.StoreInt32((*int32)(s), int32(val))
}

// String is part of the flag.Value interface.
func (s *severity) String() string {
	return strconv.FormatInt(int64(s.get()), 10)
}

// Get is part of the flag.Value interface.
func (s *severity) Get() interface{} {
	return s.get()
}

// Set is part of the flag.Value interface.
func (s *severity) Set(value string) error {
	var threshold severity

	key := strings.ToUpper(value)
	threshold, ok := severityByName[key]

	// If it is not a known name, then parse it as an int.
	if !ok {
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}

		threshold = severity(v)
	}

	s.set(threshold)
	return nil
}

// Level is exported because it appears in the arguments to V and is
// the type of the --verbosity flag, which can be set programmatically.
// It's a distinct type because we want to discriminate it from logType.
// Variables of type level are only changed under logging.mu.
// The --verbosity flag is read only with atomic ops, so the state of the logging
// module is consistent.

// Level is treated as a sync/atomic int32.

// Level specifies a level of verbosity for V logs. *Level implements
// flag.Value; the --verbosity flag is of type Level and should be modified
// only through the flag.Value interface.
type Level int32 // sync/atomic int32

// get returns the value of the Level.
func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

// set sets the value of the Level.
func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

// String is part of the flag.Value interface.
func (l *Level) String() string {
	return strconv.FormatInt(int64(l.get()), 10)
}

// Get is part of the flag.Value interface.
func (l *Level) Get() interface{} {
	return l.get()
}

// Set is part of the flag.Value interface.
func (l *Level) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	l.set(Level(v))
	return nil
}

// modulePat contains a filter for the --vmodule flag.
// It holds a verbosity level and a file pattern to match.
type modulePat struct {
	pattern string
	literal bool // The pattern is a literal string
	level   Level
}

// match reports whether the file matches the pattern. It uses a string
// comparison if the pattern contains no metacharacters.
func (m *modulePat) match(file string) bool {
	if m.literal {
		return file == m.pattern
	}

	match, _ := filepath.Match(m.pattern, file)
	return match
}

// moduleSpec represents the setting of the --vmodule flag.
type moduleSpec struct {
	sync.RWMutex

	set     int32
	filters []modulePat
	vmap    map[uintptr]Level
}

func (m *moduleSpec) isSet() bool {
	return atomic.LoadInt32(&m.set) > 0
}

func (m *moduleSpec) getV(pc uintptr) Level {
	m.RLock()
	v, ok := m.vmap[pc]
	m.RUnlock()

	if ok {
		return v
	}

	return m.setV(pc)
}

// setV computes and memoizes the V level for a given PC when vmodule is enabled.
// File pattern matching takes the basename of the file, stripped of its .go suffix,
// and uses filepath.Match, which is a little more general than the *? matching in C++.
func (m *moduleSpec) setV(pc uintptr) Level {
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()

	// The file is something like /a/b/c/d.go.
	// We just want the d.
	file := strings.TrimSuffix(frame.File, ".go")

	if slash := strings.LastIndexByte(file, '/'); slash >= 0 {
		file = file[slash+1:]
	}

	m.Lock()
	defer m.Unlock()

	if v, ok := m.vmap[pc]; ok {
		// someone did the work for us while we were waiting on the lock.
		return v
	}

	for _, filter := range m.filters {
		if filter.match(file) {
			m.vmap[pc] = filter.level
			return filter.level
		}
	}

	m.vmap[pc] = 0
	return 0
}

func (m *moduleSpec) String() string {
	// Lock because the type is not atomic. TODO: clean this up.
	m.RLock()
	defer m.RUnlock()

	b := new(bytes.Buffer)
	for i, f := range m.filters {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(f.pattern)
		b.WriteByte('=')
		b.WriteString(strconv.Itoa(int(f.level)))
	}

	return b.String()
}

// Get is part of the (Go 1.2) flag.Getter interface.
// It always returns nil for this flag type since the struct is not exported.
func (m *moduleSpec) Get() interface{} {
	return nil
}

var errVmoduleSyntax = errors.New("syntax error: expect comma-separated list of filename=N")

// Syntax: --vmodule=recordio=2,file=1,gfs*=3
func (m *moduleSpec) Set(value string) error {
	var filters []modulePat

	for _, pat := range strings.Split(value, ",") {
		if len(pat) == 0 {
			// Empty strings such as from a trailing comma can be ignored.
			continue
		}

		patLev := strings.Split(pat, "=")
		if len(patLev) != 2 || len(patLev[0]) == 0 || len(patLev[1]) == 0 {
			return errVmoduleSyntax
		}

		pattern := patLev[0]

		v, err := strconv.Atoi(patLev[1])
		if err != nil {
			return errors.New("syntax error: expect comma-separated list of filename=N")
		}

		if v < 0 {
			return errors.New("negative value for vmodule level")
		}
		if v == 0 {
			continue // Ignore. It's harmless but no point in paying the overhead.
		}

		literal, err := isLiteral(pattern)
		if err != nil {
			return err
		}

		// TODO: check syntax of filter?
		filters = append(filters, modulePat{
			pattern: pattern,
			literal: literal,
			level:   Level(v),
		})
	}

	m.filters = filters
	m.vmap = make(map[uintptr]Level)
	atomic.StoreInt32(&m.set, int32(len(filters)))

	return nil
}

// isLiteral reports whether the pattern is a literal string,
// that is, has no metacharacters that require filepath.Match to be called to match the pattern.
//
// The error returned should be filepath.ErrBadPattern,
// but filepath should provide the check pattern for syntax, which it doesn’t.
func isLiteral(pattern string) (bool, error) {
	return !strings.ContainsAny(pattern, `\*?[]`), nil
}

// traceLocation represents the setting of the --log_backtrace_at flag.
type traceLocation struct {
	sync.RWMutex

	file string
	line int32
}

// isSet reports whether the trace location has been specified.
// logging.mu is held.
func (t *traceLocation) isSet() bool {
	return atomic.LoadInt32(&t.line) > 0
}

// match reports whether the specified file and line matches the trace location.
// The argument file name is the full path, not the basename specified in the flag.
// logging.mu is held.
func (t *traceLocation) match(file string, line int) bool {
	// Lock because the type is not atomic. TODO: clean this up.
	t.RLock()
	defer t.RUnlock()

	if atomic.LoadInt32(&t.line) != int32(line) {
		return false
	}

	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}

	return t.file == file
}

func (t *traceLocation) String() string {
	// Lock because the type is not atomic. TODO: clean this up.
	t.RLock()
	defer t.RUnlock()

	return fmt.Sprintf("%s:%d", t.file, atomic.LoadInt32(&t.line))
}

// Get is part of the (Go 1.2) flag.Getter interface.
// It always returns nil for this flag type since the struct is not exported.
func (t *traceLocation) Get() interface{} {
	return nil
}

var errTraceSyntax = errors.New("syntax error: expect file.go:234")

// Syntax: --log_backtrace_at=gopherflakes.go:234
// Note that unlike vmodule the file extension is included here.
func (t *traceLocation) Set(value string) error {
	t.Lock()
	defer t.Unlock()

	if value == "" {
		// Unset.
		t.line = 0
		t.file = ""

		return nil
	}

	fields := strings.Split(value, ":")
	if len(fields) != 2 {
		return errTraceSyntax
	}

	file, line := fields[0], fields[1]
	if !strings.Contains(file, ".") {
		return errTraceSyntax
	}

	v, err := strconv.Atoi(line)
	if err != nil {
		return errTraceSyntax
	}

	if v <= 0 {
		return errors.New("negative or zero value for level")
	}

	t.file = file
	atomic.StoreInt32(&t.line, int32(v))

	return nil
}

func init() {
	flag.Struct("", &logging.flagT)
}

// flagT collects all the flags of the logging setup.
type flagT struct {
	// Boolean flags. NOT ATOMIC or thread-safe.
	ToStderr     bool `flag:"logtostderr"     desc:"log to standard error instead of files"`
	AlsoToStderr bool `flag:"alsologtostderr" desc:"log to standard error as well as files"`

	// Level flag. Handled atomically.
	StderrThreshold severity `flag:"stderrthreshold,def=ERROR" desc:"logs at or above this ·threshold· go to stderr"`

	// traceLocation is the state of the --log_backtrace_at flag.
	TraceLocation traceLocation `flag:"log_backtrace_at" desc:"when logging hits line ·file:N·, emit a stack trace"`
	// These flags are modified only under lock,
	// although verbosity may be fetched safely using atomic.LoadInt32.
	Vmodule   moduleSpec `flag:"vmodules"  desc:"comma-separated list of ·pattern=N· settings for file-filtered logging"`
	Verbosity Level      `flag:"verbosity" desc:"log ·level· for V logs"`
}
