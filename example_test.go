// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package log

import (
	"bytes"
	"fmt"
)

func ExampleLogger() {
	var buf bytes.Buffer

	logger := New(&buf, Lshortfile, LevelWarning|LevelError|LevelFatal)

	logger.Print("print")
	logger.Printf("%s\n", "printf")
	logger.Println("println")

	logger.Debug("debug")           // ignored.
	logger.Debugf("%s\n", "debugf") // ignored.
	logger.Debugln("debugln")       // ignored.

	logger.Info("info")           // ignored.
	logger.Infof("%s\n", "infof") // ignored.
	logger.Infoln("infoln")       // ignored.

	logger.Warning("warning")
	logger.Warningf("%s\n", "warningf")
	logger.Warningln("warningln")

	logger.Error("error")
	logger.Errorf("%s\n", "errorf")
	logger.Errorln("errorln")

	logger.Fatal("fatal")
	logger.Fatalf("%s\n", "fatalf")
	logger.Fatalln("fatalln")

	fmt.Print(&buf)
	// Output:
	// example_test.go:17: print
	// example_test.go:18: printf
	// example_test.go:19: println
	// WARN example_test.go:29: warning
	// WARN example_test.go:30: warningf
	// WARN example_test.go:31: warningln
	// ERRO example_test.go:33: error
	// ERRO example_test.go:34: errorf
	// ERRO example_test.go:35: errorln
	// FATA example_test.go:37: fatal
	// FATA example_test.go:38: fatalf
	// FATA example_test.go:39: fatalln
}

func Example() {
	var buf bytes.Buffer

	SetLevels(LevelWarning | LevelError | LevelFatal)
	SetFlags(Lshortfile)
	SetOutput(&buf)

	Print("print")
	Printf("%s\n", "printf")
	Println("println")

	Debug("debug")           // ignored.
	Debugf("%s\n", "debugf") // ignored.
	Debugln("debugln")       // ignored.

	Info("info")           // ignored.
	Infof("%s\n", "infof") // ignored.
	Infoln("infoln")       // ignored.

	Warning("warning")
	Warningf("%s\n", "warningf")
	Warningln("warningln")

	Error("error")
	Errorf("%s\n", "errorf")
	Errorln("errorln")

	fmt.Print(&buf)
	// Output:
	// example_test.go:64: print
	// example_test.go:65: printf
	// example_test.go:66: println
	// WARN example_test.go:76: warning
	// WARN example_test.go:77: warningf
	// WARN example_test.go:78: warningln
	// ERRO example_test.go:80: error
	// ERRO example_test.go:81: errorf
	// ERRO example_test.go:82: errorln
}
