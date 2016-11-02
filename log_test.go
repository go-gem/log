// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.
package log_test

import (
	"bytes"
	"fmt"
	"github.com/go-gem/log"
)

func ExampleLogger() {
	var buf bytes.Buffer

	var logger log.Logger
	logger = log.New(&buf, log.Lshortfile, log.LevelWarning|log.LevelError|log.LevelFatal)

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
	// WARNING: log_test.go:26: warning
	// WARNING: log_test.go:27: warningf
	// WARNING: log_test.go:28: warningln
	// ERROR: log_test.go:30: error
	// ERROR: log_test.go:31: errorf
	// ERROR: log_test.go:32: errorln
	// FATAL: log_test.go:34: fatal
	// FATAL: log_test.go:35: fatalf
	// FATAL: log_test.go:36: fatalln
}
