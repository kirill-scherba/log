// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// LogEntry is a struct that represents a log entry.
//
// It contains the following fields:
//
//   - Timestamp: the timestamp for the log entry.
//   - Level: the log level for the log entry.
//   - Message: the log message for the log entry.
//   - Fields: a map of additional fields to be included in the log
//     entry.
//
// The String method returns a JSON representation of the log entry.
type LogEntry struct {
	AppType   string         `json:"app_type"` // Prod, Dev, Test or some other
	Timestamp string         `json:"@timestamp"`
	Level     LogLevel       `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// LogLevel represents a log level.
type LogLevel string

// String returns a string representation of a log entry.
//
// It formats the log entry as a string in the following format:
//
//	<timestamp> [<level>] <message>[, fields: <fields>]
//
// The timestamp is formatted as per RFC3339. The level is the log level.
// The message is the log message. The fields are the additional fields that
// were passed in when creating the log entry.
func (entry *LogEntry) String() string {
	// Format the log entry as a string.

	// If the fields are not empty, format them as a string.
	var fields string
	if entry.Fields != nil {
		fields = fmt.Sprintf(", fields: %v", entry.Fields)
	}

	// If the level is not none, format it as a string.
	var level string
	if entry.Level != LevelNone {
		level = "[" + string(entry.Level) + "] "
	}

	// Return the formatted log entry as a string.
	return fmt.Sprintf(
		`%-36s %s%s%s`,
		entry.Timestamp, level, strings.Trim(entry.Message, "\n"), fields,
	)
}

// Json returns a JSON representation of a log entry.
//
// It marshals the log entry into a JSON byte slice and returns the JSON
// representation of the log entry as a string.
func (entry *LogEntry) Json() string {

	// Remove trailing newline from message
	entry.Message = strings.Trim(entry.Message, "\n")

	// Marshal the log entry into a JSON byte slice.
	logJSON, err := json.Marshal(entry)
	if err != nil {
		// Log and exit if we fail to marshal the log entry.
		return fmt.Sprintf(
			`{"error": "Failed to marshal log entry: %v"}`, err,
		)
	}

	// Return the JSON representation of the log entry as a string.
	return string(logJSON)
}

// entry returns a log entry with the given level, message, and fields.
// It is a convenience function for creating log entries.
//
// The fields parameter is a variable argument list of maps, allowing
// the caller to pass in any number of fields to be included in the log
// entry. The first map in the list is used as the fields for the log
// entry.
func entry(level LogLevel, v ...any) *LogEntry {

	// Get fields map[string]any from last element of v and remove it from v
	v, fields := getFields(v)

	// Make message string from v
	message := fmt.Sprint(v...)

	return &LogEntry{
		AppType:   appType,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   message,
		Level:     LogLevel(level),
		Fields:    fields,
	}
}

// entryf returns a log entry with the given level, format string, and values.
// It is a convenience function for creating log entries.
//
// The format string is used to format the values passed in via the variable argument
// list. The resulting log entry will contain the formatted message and the fields passed
// in via the variable argument list.
//
// The fields parameter is a variable argument list of maps, allowing the caller to pass
// in any number of fields to be included in the log entry. The first map in the list is
// used as the fields for the log entry.
func entryf(level LogLevel, format string, v ...any) *LogEntry {
	// Get fields map[string]any from last element of v and remove it from v
	v, fields := getFields(v)

	// Return a log entry with the given level, message, and fields
	return entry(level, fmt.Sprintf(format, v...), fields)
}

// getFields takes a variable argument list of values and returns a slice of the
// original values and a map[string]any that contains the fields of the last
// value in the list. If the list is empty, it returns an empty slice and a nil
// map.
//
// The last element of v is expected to be a map[string]any
func getFields(v []any) (vout []any, fields Fields) {

	// By default, vout equals v
	vout = v

	// Return empty slice vout and nil map if v is empty
	if len(v) == 0 {
		return
	}

	// Get fields map[string]any or Fields from last element of v and remove
	// it from v if it exists. The last element of v is expected to be a Fields
	switch f := v[len(v)-1].(type) {
	case map[string]any:
		fields, vout = f, v[:len(v)-1]
	case Fields:
		fields, vout = f, v[:len(v)-1]
	}

	return
}
