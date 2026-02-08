// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"log"
	"testing"
	"time"
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
	Init(&Config{AppShort: "log-test", AppType: "DEV", UseStdout: true,
		FileConfig: &FileConfig{Folder: "/tmp"},
	})
	defer Close()

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

func TestLogger(t *testing.T) {

	t.Run("Init", func(t *testing.T) {

		// Initialise logger
		Init(&Config{AppShort: "log-test", AppType: "DEV", UseStdout: true,
			FileConfig: &FileConfig{
				Folder: "/tmp", 
				CreateNewAfter: 1 * time.Second,
				RemoveOldAfter: 1 * time.Minute,
				RemoveSuffixes: []string{".log", ".log.gz"},
			},
		})
		defer Close()

		// Send a message to the default logger
		Debug("some debug message 1", map[string]any{"key": "value"})

		// Sleep 2 seconds to wait for the log file name to be renewated
		time.Sleep(2 * time.Second)

		// Send a message to the default logger
		Debug("some debug message 2", map[string]any{"key": "value"})

		// time.Sleep(3 * time.Second)
	})
}
