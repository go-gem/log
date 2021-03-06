// Copyright (c) 2009 The Go Authors and 2016, Gem Authors
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

// These tests are too simple.

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	Rdate         = `[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9]`
	Rtime         = `[0-9][0-9]:[0-9][0-9]:[0-9][0-9]`
	Rmicroseconds = `\.[0-9][0-9][0-9][0-9][0-9][0-9]`
	Rline         = `(56|58):` // must update if the calls to l.Printf / l.Print below move
	Rlongfile     = `.*/[A-Za-z0-9_\-]+\.go:` + Rline
	Rshortfile    = `[A-Za-z0-9_\-]+\.go:` + Rline
)

type tester struct {
	flag    int
	pattern string // regexp that log output must match; we add ^ and expected_text$ always
}

var tests = []tester{
	// individual pieces:
	{0, ""},
	{0, ""},
	{Ldate, Rdate + " "},
	{Ltime, Rtime + " "},
	{Ltime | Lmicroseconds, Rtime + Rmicroseconds + " "},
	{Lmicroseconds, Rtime + Rmicroseconds + " "}, // microsec implies time
	{Llongfile, Rlongfile + " "},
	{Lshortfile, Rshortfile + " "},
	{Llongfile | Lshortfile, Rshortfile + " "}, // shortfile overrides longfile
	// everything at once:
	{Ldate | Ltime | Lmicroseconds | Llongfile, Rdate + " " + Rtime + Rmicroseconds + " " + Rlongfile + " "},
	{Ldate | Ltime | Lmicroseconds | Lshortfile, Rdate + " " + Rtime + Rmicroseconds + " " + Rshortfile + " "},
}

// Test using Println("hello", 23, "world") or using Printf("hello %d world", 23)
func testPrint(t *testing.T, flag int, pattern string, useFormat bool) {
	buf := new(bytes.Buffer)
	SetOutput(buf)
	SetFlags(flag)
	if useFormat {
		Printf("hello %d world", 23)
	} else {
		Println("hello", 23, "world")
	}
	line := buf.String()
	line = line[0 : len(line)-1]
	pattern = "^" + pattern + "hello 23 world$"
	matched, err4 := regexp.MatchString(pattern, line)
	if err4 != nil {
		t.Fatal("pattern did not compile:", err4)
	}
	if !matched {
		t.Errorf("log output should match %q is %q", pattern, line)
	}
	SetOutput(os.Stderr)
}

func TestAll(t *testing.T) {
	for _, testcase := range tests {
		testPrint(t, testcase.flag, testcase.pattern, false)
		testPrint(t, testcase.flag, testcase.pattern, true)
	}
}

func TestOutput(t *testing.T) {
	const testString = "test"
	var b bytes.Buffer
	l := New(&b, 0, LevelAll)
	l.Println(testString)
	if expect := testString + "\n"; b.String() != expect {
		t.Errorf("log output should match %q is %q", expect, b.String())
	}
}

func TestFlagAndPrefixSetting(t *testing.T) {
	var b bytes.Buffer
	l := New(&b, LstdFlags, LevelAll)
	f := l.Flags()
	if f != LstdFlags {
		t.Errorf("Flags 1: expected %x got %x", LstdFlags, f)
	}
	l.SetFlags(f | Lmicroseconds)
	f = l.Flags()
	if f != LstdFlags|Lmicroseconds {
		t.Errorf("Flags 2: expected %x got %x", LstdFlags|Lmicroseconds, f)
	}
	// Verify a log message looks right, with our prefix and microseconds present.
	l.Print("hello")
	pattern := "^" + Rdate + " " + Rtime + Rmicroseconds + " hello\n"
	matched, err := regexp.Match(pattern, b.Bytes())
	if err != nil {
		t.Fatalf("pattern %q did not compile: %s", pattern, err)
	}
	if !matched {
		t.Error("message did not match pattern")
	}
}

func TestUTCFlag(t *testing.T) {
	var b bytes.Buffer
	l := New(&b, LstdFlags, LevelAll)
	l.SetFlags(Ldate | Ltime | LUTC)
	// Verify a log message looks right in the right time zone. Quantize to the second only.
	now := time.Now().UTC()
	l.Print("hello")
	want := fmt.Sprintf("%d/%.2d/%.2d %.2d:%.2d:%.2d hello\n",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	got := b.String()
	if got == want {
		return
	}
	// It's possible we crossed a second boundary between getting now and logging,
	// so add a second and try again. This should very nearly always work.
	now = now.Add(time.Second)
	want = fmt.Sprintf("%d/%.2d/%.2d %.2d:%.2d:%.2d hello\n",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	if got == want {
		return
	}
	t.Errorf("got %q; want %q", got, want)
}

func TestEmptyPrintCreatesLine(t *testing.T) {
	var b bytes.Buffer
	l := New(&b, LstdFlags, LevelAll)
	l.Print()
	l.Println("non-empty")
	output := b.String()
	if n := strings.Count(output, "\n"); n != 2 {
		t.Errorf("expected 2 lines, got %d", n)
	}
}

func TestSetFlags(t *testing.T) {
	flag := Lshortfile
	SetFlags(flag)
	if flag != Flags() {
		t.Errorf("expected falg %d, got %d", flag, Flags())
	}
}

func TestSetLevels(t *testing.T) {
	level := LevelDebug
	SetLevels(level)
	if level != Levels() {
		t.Errorf("expected level %d, got %d", level, Levels())
	}
}

func BenchmarkItoa(b *testing.B) {
	dst := make([]byte, 0, 64)
	for i := 0; i < b.N; i++ {
		dst = dst[0:0]
		itoa(&dst, 2015, 4)   // year
		itoa(&dst, 1, 2)      // month
		itoa(&dst, 30, 2)     // day
		itoa(&dst, 12, 2)     // hour
		itoa(&dst, 56, 2)     // minute
		itoa(&dst, 0, 2)      // second
		itoa(&dst, 987654, 6) // microsecond
	}
}

func BenchmarkPrintln(b *testing.B) {
	const testString = "test"
	var buf bytes.Buffer
	l := New(&buf, LstdFlags, LevelAll)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		l.Println(testString)
	}
}
