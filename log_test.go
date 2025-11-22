// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"log"
	"testing"
)

func TestLog(t *testing.T) {

	// Test entry
	t.Log(entry(LevelDebug, "entry() test", map[string]any{"key": "value"}))

	// Test SDebug
	t.Log(Sdebug("Sdebug() test", map[string]any{"key": "value"}))

	// Test Log
	PrintLevel(LevelDebug, "PrintLevel() test")

	// Test Log Info
	Info("Info() test", map[string]any{"key": "value"})

	// Test Log Debugf
	Debugf("Debugf() test %d", 48)
	Debugf("Debugf() test %d with fields", 48, map[string]any{"key": "value"})

	// Initialise logger to check default log print
	Init(Config{AppShort: "log-test", AppType: "DEV", UseStdout: true,
		FileConfig: &FileConfig{Folder: "/tmp"},
	})
	defer CLose()

	// Test default log print
	log.Println("log.Println() (default log) test")

	// Test default log print when DefaultLevel is INFO
	SetDefaultLevel(LevelInfo)
	log.Println("log.Println() (default log) test")

	// Test default log print when DefaultLevel is NONE. This message should not
	// be printed
	SetDefaultLevel(LevelNone)
	log.Println("log.Println() (default log) test")

	// Some debug message with default log level set to NONE
	Debug("some debug message", map[string]any{"key": "value"})
}
