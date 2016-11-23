// Copyright (c) 2009 The Go Authors and 2016, Gem Authors
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package log implements a simple leveled logging package. It defines a type,
// Logger, with methods for formatting output. It also has a predefined 'standard'
// Logger accessible through helper functions Print[f|ln], Fatal[f|ln], and
// Panic[f|ln], which are easier to use than creating a Logger manually.
// That logger writes to standard error and prints the date and time
// of each logged message.
// The Fatal functions call os.Exit(1) after writing the log message.
// The Panic functions call panic after writing the log message.
package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// These flags define which text to prefix to each log entry generated by the Logger.
const (
	// Bits or'ed together to control what's printed.
	// There is no control over the order they appear (the order listed
	// here) or the format they present (as described in the comments).
	// The prefix is followed by a colon only when Llongfile or Lshortfile
	// is specified.
	// For example, flags Ldate | Ltime (or LstdFlags) produce,
	//	2009/01/23 01:23:23 message
	// while flags Ldate | Ltime | Lmicroseconds | Llongfile produce,
	//	2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

// levels.
// LevelAll should not be using with other levels at the same time,
// unless you know what are you doing.
const (
	LevelDebug = 1 << iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal

	LevelAll = LevelDebug | LevelInfo | LevelWarning | LevelError | LevelFatal
)

// levels string.
const (
	prefixEmpty   = ""
	prefixDebug   = "DEBU "
	prefixInfo    = "INFO "
	prefixWarning = "WARN "
	prefixError   = "ERRO "
	prefixFatal   = "FATA "
)

// ignore return bool indicate whether the current level's log should be ignored.
func (l *Logger) ignore(level int) bool {
	return (l.level & level) == 0
}

// A Logger represents an active logging object that generates lines of
// output to an io.Writer. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu    sync.Mutex // ensures atomic writes; protects the following fields
	level int        // logging level
	flag  int        // properties
	out   io.Writer  // destination for output
	buf   []byte     // for accumulating text to write
}

// New creates a new Logger. The out variable sets the
// destination to which log data will be written.
// The prefix appears at the beginning of each generated log line.
// The flag argument defines the logging properties.
func New(out io.Writer, flag, level int) *Logger {
	return &Logger{out: out, flag: flag, level: level}
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

var std = New(os.Stderr, LstdFlags, LevelAll)

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *Logger) formatHeader(buf *[]byte, prefix string, t time.Time, file string, line int) {
	*buf = append(*buf, prefix...)
	if l.flag&LUTC != 0 {
		t = t.UTC()
	}
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is used to recover the PC and is
// provided for generality, although at the moment on all pre-defined
// paths it will be 2.
func (l *Logger) Output(calldepth int, s string, prefix string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, prefix, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...), prefixEmpty)
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...), prefixEmpty)
}

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Println(v ...interface{}) {
	l.Output(2, fmt.Sprintln(v...), prefixEmpty)
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debug(v ...interface{}) {
	if l.ignore(LevelDebug) {
		return
	}
	l.Output(2, fmt.Sprint(v...), prefixDebug)
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.ignore(LevelDebug) {
		return
	}
	l.Output(2, fmt.Sprintf(format, v...), prefixDebug)
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Debugln(v ...interface{}) {
	if l.ignore(LevelDebug) {
		return
	}
	l.Output(2, fmt.Sprintln(v...), prefixDebug)
}

// Info calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	if l.ignore(LevelInfo) {
		return
	}
	l.Output(2, fmt.Sprint(v...), prefixInfo)
}

// Infof calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.ignore(LevelInfo) {
		return
	}
	l.Output(2, fmt.Sprintf(format, v...), prefixInfo)
}

// Infoln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Infoln(v ...interface{}) {
	if l.ignore(LevelInfo) {
		return
	}
	l.Output(2, fmt.Sprintln(v...), prefixInfo)
}

// Warning calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warning(v ...interface{}) {
	if l.ignore(LevelWarning) {
		return
	}
	l.Output(2, fmt.Sprint(v...), prefixWarning)
}

// Warningf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warningf(format string, v ...interface{}) {
	if l.ignore(LevelWarning) {
		return
	}
	l.Output(2, fmt.Sprintf(format, v...), prefixWarning)
}

// Warningln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Warningln(v ...interface{}) {
	if l.ignore(LevelWarning) {
		return
	}
	l.Output(2, fmt.Sprintln(v...), prefixWarning)
}

// Error calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	if l.ignore(LevelError) {
		return
	}
	l.Output(2, fmt.Sprint(v...), prefixError)
}

// Errorf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.ignore(LevelError) {
		return
	}
	l.Output(2, fmt.Sprintf(format, v...), prefixError)
}

// Errorln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Errorln(v ...interface{}) {
	if l.ignore(LevelError) {
		return
	}
	l.Output(2, fmt.Sprintln(v...), prefixError)
}

// Fatal calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Fatal(v ...interface{}) {
	if l.ignore(LevelFatal) {
		return
	}
	l.Output(2, fmt.Sprint(v...), prefixFatal)
}

// Fatalf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.ignore(LevelFatal) {
		return
	}
	l.Output(2, fmt.Sprintf(format, v...), prefixFatal)
}

// Fatalln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Fatalln(v ...interface{}) {
	if l.ignore(LevelFatal) {
		return
	}
	l.Output(2, fmt.Sprintln(v...), prefixFatal)
}

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(2, s, prefixEmpty)
	panic(s)
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Output(2, s, prefixEmpty)
	panic(s)
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Output(2, s, prefixEmpty)
	panic(s)
}

// Flags returns the output flags for the logger.
func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

// SetFlags sets the output flags for the logger.
func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

// Levels returns the levels for the logger.
func (l *Logger) Levels() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// SetLevels sets the levels for the logger.
func (l *Logger) SetLevels(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

// Flags returns the output flags for the standard logger.
func Flags() int {
	return std.Flags()
}

// SetFlags sets the output flags for the standard logger.
func SetFlags(flag int) {
	std.SetFlags(flag)
}

// Levels returns the levels for the standard logger.
func Levels() int {
	return std.Levels()
}

// SetLevels sets the levels for the standard logger.
func SetLevels(level int) {
	std.SetLevels(level)
}

// These functions write to the standard logger.

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	std.Output(2, fmt.Sprint(v...), prefixEmpty)
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(format, v...), prefixEmpty)
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...), prefixEmpty)
}

// Debug calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	if std.ignore(LevelDebug) {
		return
	}
	std.Output(2, fmt.Sprint(v...), prefixDebug)
}

// Debugf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	if std.ignore(LevelDebug) {
		return
	}
	std.Output(2, fmt.Sprintf(format, v...), prefixDebug)
}

// Debugln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Debugln(v ...interface{}) {
	if std.ignore(LevelDebug) {
		return
	}
	std.Output(2, fmt.Sprintln(v...), prefixDebug)
}

// Info calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	if std.ignore(LevelInfo) {
		return
	}
	std.Output(2, fmt.Sprint(v...), prefixInfo)
}

// Infof calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	if std.ignore(LevelInfo) {
		return
	}
	std.Output(2, fmt.Sprintf(format, v...), prefixInfo)
}

// Infoln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Infoln(v ...interface{}) {
	if std.ignore(LevelInfo) {
		return
	}
	std.Output(2, fmt.Sprintln(v...), prefixInfo)
}

// Warning calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func Warning(v ...interface{}) {
	if std.ignore(LevelWarning) {
		return
	}
	std.Output(2, fmt.Sprint(v...), prefixWarning)
}

// Warningf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Warningf(format string, v ...interface{}) {
	if std.ignore(LevelWarning) {
		return
	}
	std.Output(2, fmt.Sprintf(format, v...), prefixWarning)
}

// Warningln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Warningln(v ...interface{}) {
	if std.ignore(LevelWarning) {
		return
	}
	std.Output(2, fmt.Sprintln(v...), prefixWarning)
}

// Error calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	if std.ignore(LevelError) {
		return
	}
	std.Output(2, fmt.Sprint(v...), prefixError)
}

// Errorf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	if std.ignore(LevelError) {
		return
	}
	std.Output(2, fmt.Sprintf(format, v...), prefixError)
}

// Errorln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Errorln(v ...interface{}) {
	if std.ignore(LevelError) {
		return
	}
	std.Output(2, fmt.Sprintln(v...), prefixError)
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	if std.ignore(LevelFatal) {
		return
	}
	std.Output(2, fmt.Sprint(v...), prefixFatal)
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	if std.ignore(LevelFatal) {
		return
	}
	std.Output(2, fmt.Sprintf(format, v...), prefixFatal)
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	if std.ignore(LevelFatal) {
		return
	}
	std.Output(2, fmt.Sprintln(v...), prefixFatal)
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	std.Output(2, s, prefixEmpty)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.Output(2, s, prefixEmpty)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(2, s, prefixEmpty)
	panic(s)
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is the count of the number of
// frames to skip when computing the file name and line number
// if Llongfile or Lshortfile is set; a value of 1 will print the details
// for the caller of Output.
func Output(calldepth int, s string) error {
	return std.Output(calldepth+1, s, prefixEmpty) // +1 for this frame.
}
