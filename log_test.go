// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package log_test

import (
	"bytes"
	"github.com/go-gem/log"
	"strings"
	"testing"
)

func TestStdLogger_Fatal(t *testing.T) {
	var buf bytes.Buffer

	var logger log.Logger
	logger = log.New(&buf, log.Lshortfile, log.LevelFatal)

	defer func() {
		if rcv := recover(); rcv != nil {
			want := "FATAL: log_test.go:31: fatal\n"
			if !strings.EqualFold(buf.String(), want) {
				t.Fatalf("Unexpected result: got %q want %q", buf.String(), want)
			}
		} else {
			t.Fatal("Failed to catch panic.")
		}
	}()

	logger.Fatal("fatal")
}

func TestStdLogger_Fatalf(t *testing.T) {
	var buf bytes.Buffer

	var logger log.Logger
	logger = log.New(&buf, log.Lshortfile, log.LevelFatal)

	defer func() {
		if rcv := recover(); rcv != nil {
			want := "FATAL: log_test.go:51: fatalf\n"
			if !strings.EqualFold(buf.String(), want) {
				t.Fatalf("Unexpected result: got %q want %q", buf.String(), want)
			}
		} else {
			t.Fatal("Failed to catch panic.")
		}
	}()

	logger.Fatalf("%s\n", "fatalf")
}

func TestStdLogger_Fatalln(t *testing.T) {
	var buf bytes.Buffer

	var logger log.Logger
	logger = log.New(&buf, log.Lshortfile, log.LevelFatal)

	defer func() {
		if rcv := recover(); rcv != nil {
			want := "FATAL: log_test.go:71: fatalln\n"
			if !strings.EqualFold(buf.String(), want) {
				t.Fatalf("Unexpected result: got %q want %q", buf.String(), want)
			}
		} else {
			t.Fatal("Failed to catch panic.")
		}
	}()

	logger.Fatalln("fatalln")
}
