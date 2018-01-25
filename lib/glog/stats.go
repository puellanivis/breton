package glog

import (
	"sync/atomic"
)

// OutputStats tracks the number of output lines and bytes written.
type OutputStats struct {
	lines int64
	bytes int64
}

// Lines returns the number of lines written.
func (s *OutputStats) Lines() int64 {
	return atomic.LoadInt64(&s.lines)
}

// Bytes returns the number of bytes written.
func (s *OutputStats) Bytes() int64 {
	return atomic.LoadInt64(&s.bytes)
}

func (s *OutputStats) add(n int64) {
	atomic.AddInt64(&s.lines, 1)
	atomic.AddInt64(&s.bytes, n)
}

// Stats tracks the number of lines of output and number of bytes
// per severity level. Values must be read with atomic.LoadInt64.
var Stats struct {
	Info, Warning, Error OutputStats
}

var severityStats = []*OutputStats{
	infoLog:    &Stats.Info,
	warningLog: &Stats.Warning,
	errorLog:   &Stats.Error,
}
