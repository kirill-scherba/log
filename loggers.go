// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"slices"
	"sync"
)

// loggersType is a struct that holds information about how to send log entries to
// loggers.
type loggersType struct {

	// useStdoutLogger is a boolean that indicates whether to use the stdout logger
	useStdoutLogger bool

	// useEsLogger is a boolean that indicates whether to use the Elasticsearch logger
	useEsLogger bool

	// useFailLogger is a boolean that indicates whether to use the fail logger
	useFailLogger bool

	// filterLevels is a list of log levels to filter out.
	filterLevels []LogLevel

	// Elasticsearch logger
	*es

	// Fail logger
	*file

	// Start wait group
	wgStart sync.WaitGroup

	// Close wait group
	wgClose sync.WaitGroup
}

// newLoggers returns a new loggersType with an entry channel and two parameters set to default values.
// It also starts a goroutine that handles log entries in the entry channel.
func newLoggers() (l *loggersType) {
	// Create a new loggersType with default values
	l = &loggersType{
		useStdoutLogger: true,    // Set log to stdout by default
		es:              &es{},   // Create a new Elasticsearch logger object
		file:            &file{}, // Create a new fail logger object
	}
	return
}

// send sends a log entry to stdout logger and to the entry channel.
// The entry channel is consumed by the entryHandler goroutine, which aggregates log entries
// in a slice until either the slice reaches the maximum size (l.entriesToHold) or the time to hold
// (l.timeToHold) expires. When either condition is met, it sends the aggregated log entries to
// Elasticsearch using the sendToElasticsearch method.
func (l *loggersType) send(entry *LogEntry) (err error) {

	// Filter logger entries by level
	if l.filterLevels != nil {
		if slices.Contains(l.filterLevels, entry.Level) {
			return
		}
	}

	// Send to stdout logger. The stdout logger is a logger that writes to
	// stdout.
	if l.useStdoutLogger {
		stdoutLogger.Println(entry.String())
	}

	// Send to Elasticsearch channel which will send to elasticsearch
	// The entry channel is a channel that receives log entries.
	// It is consumed by the entryHandler goroutine, which aggregates log entries in a slice until
	// either the slice reaches the maximum size (l.entriesToHold) or the time to hold (l.timeToHold) expires.
	// When either condition is met, it sends the aggregated log entries to Elasticsearch using the
	// sendToElasticsearch method.
	if l.useEsLogger {
		l.esEntryChannel <- entry
	}

	// Send to fail logger
	if l.useFailLogger {
		l.fileEntryChannel <- entry
	}

	return
}
