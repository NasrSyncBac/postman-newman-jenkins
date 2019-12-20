// Package log implements a simple log file. It is similar to the standard library log package,
// but it introduces log levels to control which log entries are actually written.
// Note that the various SetXXX() functions are not thread-safe and should be called before
// writing log entries (or at least while there are no parallel routines writing log entries).
// Package log is the successor to github.com/Syncbak-Git/logging.
package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Log is used for private logs. Do not create directly, use NewLog().
type Log struct {
	output    io.Writer
	logLevel  Level
	timestamp func() string
}

var std *Log

func init() {
	std = NewLog()
}

// Level is a logging level.
type Level uint64

// Defined log levels. See SetLogLevel for usage.
const (
	LevelDebug Level = 1 << iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
	LevelPanic
	LevelCustom
	LevelAll  = 0xFFFF
	LevelNone = 0
)

// SetLogLevel controls which log entries are actually written to the global log.
// Multiple logging levels can be combined by ORing individual
// Level values, eg. LevelDebug|LevelError will log both DEBUG and ERROR entries. Alternatively,
// specific log levels can be suppresed via XORing with LevelAll, eg. LevelAll ^ LevelDebug
// will log everything except DEBUG entries.
// The default log level is LevelAll.
func SetLogLevel(l Level) {
	std.SetLogLevel(l)
}

// SetOutput directs global log output to w. The default output is written to os.Stderr.
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// Close calls Close on the output Writer, if it is a WriteCloser, otherwise Close is a no-op.
func Close() error {
	return std.Close()
}

// SetOuputFile is a convenience function to wrap SetOutput() for writing global log entries to a file.
func SetOutputFile(f string) error {
	return std.SetOutputFile(f)
}

// SetTimestamp allows the user to replace the default RFC3339Nano timestamp string used by the global log. It
// is intended for creating deterministic test cases, but may be generally useful.
func SetTimestamp(f func() string) {
	std.SetTimestamp(f)
}

// Debug writes a DEBUG entry to the global log file.
func Debug(format string, args ...interface{}) error {
	return std.Debug(format, args...)
}

// Info writes a INFO entry to the global log file.
func Info(format string, args ...interface{}) error {
	return std.Info(format, args...)
}

// Warning writes a WARNING entry to the global log file.
func Warning(format string, args ...interface{}) error {
	return std.Warning(format, args...)
}

// Error writes an ERROR entry to the global log file.
func Error(format string, args ...interface{}) error {
	return std.Error(format, args...)
}

// Fatal writes a FATAL entry to the global log file and then exits
// via os.Exit(1).
func Fatal(format string, args ...interface{}) error {
	return std.Fatal(format, args...)
}

// Panic writes a PANIC entry to the global log file and then
// calls panic() with the log entry.
func Panic(format string, args ...interface{}) error {
	return std.Panic(format, args...)
}

// Custom writes a global log entry with a caller-supplied log level string.
func Custom(level string, format string, args ...interface{}) error {
	return std.Custom(level, format, args...)
}

// NewLog creates a private log with all log levels enabled and output to os.Stderr.
func NewLog() *Log {
	return &Log{
		output:   os.Stderr,
		logLevel: LevelAll,
		timestamp: func() string {
			return time.Now().UTC().Format(time.RFC3339Nano)
		},
	}
}

func (l *Log) SetLogLevel(ll Level) {
	l.logLevel = ll
}

func (l *Log) SetOutput(w io.Writer) {
	l.output = w
}

func (l *Log) Close() error {
	if o, ok := l.output.(io.WriteCloser); ok {
		return o.Close()
	}
	return nil
}

func (l *Log) SetOutputFile(f string) error {
	w, err := os.OpenFile(f, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.SetOutput(w)
	return nil
}

func (l *Log) SetTimestamp(f func() string) {
	l.timestamp = f
}

func (l *Log) Debug(format string, args ...interface{}) error {
	if l.logLevel&LevelDebug == 0 {
		return nil
	}
	return l.writeEntry("DEBUG", format, args...)
}

func (l *Log) Info(format string, args ...interface{}) error {
	if l.logLevel&LevelInfo == 0 {
		return nil
	}
	return l.writeEntry("INFO", format, args...)
}

func (l *Log) Warning(format string, args ...interface{}) error {
	if l.logLevel&LevelWarning == 0 {
		return nil
	}
	return l.writeEntry("WARNING", format, args...)
}

func (l *Log) Error(format string, args ...interface{}) error {
	if l.logLevel&LevelError == 0 {
		return nil
	}
	return l.writeEntry("ERROR", format, args...)
}

func (l *Log) Fatal(format string, args ...interface{}) error {
	if l.logLevel&LevelFatal == 0 {
		return nil
	}
	err := l.writeEntry("FATAL", format, args...)
	os.Exit(1)
	return err // won't actually execute
}

func (l *Log) Panic(format string, args ...interface{}) error {
	if l.logLevel&LevelPanic == 0 {
		return nil
	}
	err := l.writeEntry("PANIC", format, args...)
	panic(fmt.Sprintf(format, args...))
	return err // won't actually execute
}

func (l *Log) Custom(level string, format string, args ...interface{}) error {
	if l.logLevel&LevelCustom == 0 {
		return nil
	}
	return l.writeEntry(level, format, args...)
}

func (l *Log) writeEntry(level string, format string, args ...interface{}) error {
	_, err := fmt.Fprintf(l.output, "%s\t%s\t%s\n", l.timestamp(), level, fmt.Sprintf(format, args...))
	return err
}
