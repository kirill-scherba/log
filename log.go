// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package log contains log functions to send logs to stdout, files and
// Elasticsearch.
//
// The log compatible with standard log package, f.e. log.Println are available.
//
// The log messages sends to elasticsearch has json format with basic fields
// (timestamp, level, message) and additional fields in key-value format.
package log

import (
	"io"
	"log"
	"os"
	"strings"
)

// Log levels
const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelNone  LogLevel = ""
)

// Default log level uset in standart log calls, f.e. log.Println
var LevelDefault = LevelDebug

// Config is a struct that holds configuration information for the logger.
type Config struct {

	// AppShort is the short name of the application
	AppShort string

	// AppType is the type of the application, f.e. "DEV" or "PROD"
	AppType string

	// UseStdout is a boolean that indicates whether to write log messages to
	// stdout
	UseStdout bool

	// EsConfig is the configuration for the Elasticsearch logger.
	// If nil, the Elasticsearch logger is not used
	*EsConfig

	// File is the configuration for the file logger.
	// If nil, the file logger is not used
	*FileConfig

	// When loger initialized it prints "logger initialized" message. If set
	// this field to true, this message will not be printed.
	DoesNotShowInitMessage bool

	// CustomLogers is some custom loggers, it output will be parsed and added
	// to this loger output.
	CustomLogers []*log.Logger

	// FilterLevel is a list of log levels to filter out.
	FilterLevels []LogLevel
}

// Fields is a map of string to any
type Fields map[string]any

// Log level by appType
var appType string = "TEST"

// The stdoutLogger is a logger that writes to stdout
var stdoutLogger = log.New(os.Stdout, "", 0)

// loggers is a struct that holds information about how to send log entries to
// loggers.
var loggers = newLoggers()

// customWriter is io.Writer interface
type customWriter struct{}

// Write implements io.Writer interface and is used to write log entries to stdout.
// It removes timestamp from the log entry and then sends the log entry to the loggers.
// The log level is determined by the text inside `[...]` if `[` exists at the beginning of the log entry.
// If the log level is not specified, it defaults to LevelDebug.
// The log entry is then sent to the loggers, which writes it to stdout and/or to Elasticsearch.
// The Write function returns the number of bytes written and a nil error.
func (cw *customWriter) Write(p []byte) (n int, err error) {

	// If LevelDefault is LevelNone, message is ignored
	if LevelDefault == LevelNone {
		return len(p), nil
	}

	// Remove timestamp from p, p is like this '2025/10/19 10:28:50.567024 ...'
	p = p[strings.Index(string(p), " ")+1:]
	p = p[strings.Index(string(p), " ")+1:]

	// Get text inside `[...]` if `[`` exists at beginning of p
	level := string(LevelDefault)
	if strings.HasPrefix(string(p), "[") && strings.Contains(string(p), "]") {
		level = string(p[strings.Index(string(p), "[")+1 : strings.Index(string(p), "]")])
		p = p[strings.Index(string(p), "]")+1:]
	}

	loggers.send(entry(LogLevel(level), strings.TrimSpace(string(p))))

	return len(p), nil
}

// Init sets up the loggers and sets the output for the default application
// logger and for some additional loggers.
// It is called once when the application starts.
// apptype is the application type, which is used to prefix log messages.
// useStdout is a boolean that indicates whether to write log messages to stdout.
// esConfig is a pointer to an EsConfig struct, which holds information about
// how to send log entries to Elasticsearch.
// If esConfig is not nil, the loggers are set up to write log entries to Elasticsearch.
// The Elasticsearch logger config is set to the provided esConfig.
// The Elasticsearch logger config is set to the provided esConfig.
// If the MaxOverflowBuffer is 0, it is set to 2x the entryChannel size, or a
// fixed number if entryChannel size is 0.
// If the MaxRetryBuffer is 0, it is set to 100 batches.
// The Elasticsearch logger config is set to the provided esConfig.
// The Elasticsearch logger config is set to the provided esConfig.
// The loggers are then started and the output for the default application logger
// and for some additional loggers is set to a customWriter, which writes log
// entries to stdout and/or to Elasticsearch.
// Finally, a message is printed to indicate that the loggers have been initialized.
func Init(config Config) {

	// Set application type
	appType = config.AppType

	// Set useStdout
	loggers.useStdoutLogger = config.UseStdout

	// Set filter level
	loggers.filterLevels = config.FilterLevels

	// Set output for default application logger
	w := &customWriter{}
	log.SetOutput(w)

	// Add custom logers
	for _, customLogger := range config.CustomLogers {
		customLogger.SetOutput(w)
	}

	// Set elasticsearch logger config and start elasticsearch logger handler
	if config.EsConfig != nil {
		loggers.es.init(config.AppShort, config.EsConfig)
		loggers.useEsLogger = true
	}

	// Set file logger config and start file logger handler
	if config.FileConfig != nil {
		loggers.file.init(config.AppShort, config.FileConfig)
		loggers.useFailLogger = true
	}

	// Wait for loggers to start
	loggers.wgStart.Wait()

	// Print message when loggers are initialized
	if !config.DoesNotShowInitMessage {
		log.Println("logger initialized")
	}
}

// CLose closes the Elasticsearch logger and the file logger.
// It is called once when the application exits.
// It stops the Elasticsearch logger and the file logger from writing log
// entries to Elasticsearch and/or to disk.
func CLose() {
	if loggers.useEsLogger {
		loggers.es.close()
	}

	if loggers.useFailLogger {
		loggers.file.close()
	}

	loggers.wgClose.Wait()
}

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// SetDefaultLevel sets the default log level for the logger.
//
// The default log level is DEBUG. The default log level can be changed using
// this function.
//
// The default log level used in the standart log calls, f.e. log.Println.
// If the default log level is set to NONE, the logger will not write any the
// standart log calls.
//
// It takes a logLevel value as its argument, which can be any of the following:
// DEBUG, INFO, WARN, ERROR, or NONE.
//
// The logger will write log messages with the specified log level and above
// (i.e., if the log level is set to WARN, the logger will write log messages
// with the log levels WARN and ERROR).
func SetDefaultLevel(level LogLevel) {
	LevelDefault = level
}

// Sentry is a convenience function for creating log entries at the given log level.
// It takes a message, and a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as the fields for the log entry.
func Sentry(level LogLevel, v ...any) string {
	// Return a log entry with the given message and fields at the given log level.
	return entry(level, v...).String()
}

// Sentryf is a convenience function for creating log entries at the given log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
//
// The format string is used to format the values passed in via the variable argument
// list. The resulting log entry will contain the formatted message and the fields passed
// in via the variable argument list.
func Sentryf(level LogLevel, format string, v ...any) string {
	// Return a log entry with the given format string and values at the given log level.
	return entryf(level, format, v...).String()
}

// Sdebug is a convenience function for creating log entries at the debug log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Sdebug(v ...any) string {
	// Return a log entry with the given message and fields at the debug log level.
	return entry(LevelDebug, v...).String()
}

// Sdebugf is a convenience function for creating log entries at the debug log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
func Sdebugf(format string, v ...any) string {
	return entryf(LevelDebug, format, v...).String()
}

// Sinfo is a convenience function for creating log entries at the info log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Sinfo(message string, v ...any) string {
	// Return a log entry with the given message and fields at the info log level.
	return entry(LevelInfo, v...).String()
}

// Sinfof is a convenience function for creating log entries at the info log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
//
// The format string is used to format the values passed in via the variable argument
// list. The resulting log entry will contain the formatted message and the fields passed
// in via the variable argument list.
func Sinfof(format string, v ...any) string {
	// Return a log entry with the given format string and values at the info log level.
	return entryf(LevelInfo, format, v...).String()
}

// Swarn is a convenience function for creating log entries at the warn log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Swarn(message string, v ...any) string {
	// Return a log entry with the given message and fields at the warn log level.
	return entry(LevelWarn, v...).String()
}

// Swarnf is a convenience function for creating log entries at the warn log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
//
// The format string is used to format the values passed in via the variable argument
// list. The resulting log entry will contain the formatted message and the fields passed
// in via the variable argument list.
func Swarnf(format string, v ...any) string {
	// Return a log entry with the given format string and values at the warn log level.
	return entryf(LevelWarn, format, v...).String()
}

// Serror is a convenience function for creating log entries at the error log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
//
// The function returns a JSON representation of the log entry as a string.
func Serror(message string, v ...any) string {
	// Return a log entry with the given message and fields at the error log level.
	return entry(LevelError, v...).String()
}

// Serrorf is a convenience function for creating log entries at the error log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
//
// The format string is used to format the values passed in via the variable argument
// list. The resulting log entry will contain the formatted message and the fields passed
// in via the variable argument list.
func Serrorf(format string, v ...any) string {
	// Return a log entry with the given format string and values at the error log level.
	return entryf(LevelError, format, v...).String()
}

// PrintLevel is a convenience function for creating log entries at the given log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func PrintLevel(level LogLevel, v ...any) {
	loggers.send(entry(level, v...)) // Send to Stdout and Elasticsearch
}

// PrintLevelf is a convenience function for creating log entries at the given log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
//
// The format string is used to format the values passed in via the variable argument
// list. The resulting log entry will contain the formatted message and the fields passed
// in via the variable argument list.
func PrintLevelf(level LogLevel, format string, v ...any) {
	loggers.send(entryf(level, format, v...)) // Send to Stdout and Elasticsearch
}

// Println is a convenience function for creating log entries at the debug log level.
// It takes a variable argument list of values, allowing the caller to pass in any number
// of values to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Println(v ...any) { PrintLevel(LevelDefault, v...) }

// Printf is a convenience function for creating log entries at the debug log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
//
// The format string is used to format the values passed in via the variable argument
// list. The resulting log entry will contain the formatted message and the fields passed
// in via the variable argument list.
func Printf(format string, v ...any) { PrintLevelf(LevelDefault, format, v...) }

// Fatalln is a convenience function for creating log entries at the error log level
// and then exiting the program with a non-zero exit code.
//
// It takes a variable argument list of values, allowing the caller to pass in any number
// of values to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
//
// The function calls Fatal with the given values and then exits the program with a non-zero
// exit code.
func Fatalln(v ...any) { Fatal(v...) }

// Fatal is a convenience function for creating log entries at the error log level
// and then exiting the program with a non-zero exit code.
//
// It takes a variable argument list of values, allowing the caller to pass in any number
// of values to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
//
// The function calls Error with the given values and then exits the program with a non-zero
// exit code.
func Fatal(v ...any) { Error(v...); os.Exit(1) }

// Fatalf is a convenience function for creating log entries at the error log level
// and then exiting the program with a non-zero exit code.
//
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
//
// The function calls Errorf with the given format string and values and then exits the
// program with a non-zero exit code.
func Fatalf(format string, v ...any) { Errorf(format, v...); os.Exit(1) }

// Debug is a convenience function for creating log entries at the debug log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Debug(v ...any) { PrintLevel(LevelDebug, v...) }

// Debugf is a convenience function for creating log entries at the debug log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
func Debugf(format string, v ...any) { PrintLevelf(LevelDebug, format, v...) }

// Example usage:
// Debugf("Something happened with %v and %v", "foo", "bar", map[string]any{"foo": "bar"})

// Info is a convenience function for creating log entries at the info log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Info(v ...any) { PrintLevel(LevelInfo, v...) }

// Infof is a convenience function for creating log entries at the info log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
func Infof(format string, v ...any) { PrintLevelf(LevelInfo, format, v...) }

// Warn is a convenience function for creating log entries at the warn log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Warn(v ...any) { PrintLevel(LevelWarn, v...) }

// Warnf is a convenience function for creating log entries at the warn log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
func Warnf(format string, v ...any) { PrintLevelf(LevelWarn, format, v...) }

// Error is a convenience function for creating log entries at the error log level.
// It takes a variable argument list of maps, allowing the caller to pass in any number
// of fields to be included in the log entry. The first map in the list is used as
// the fields for the log entry.
func Error(v ...any) { PrintLevel(LevelError, v...) }

// Errorf is a convenience function for creating log entries at the error log level.
// It takes a format string and a variable argument list of values, allowing the caller
// to pass in any number of values to be included in the log entry. The last value in the
// list is expected to be a map[string]any, which is used as the fields for the log
// entry.
func Errorf(format string, v ...any) { PrintLevelf(LevelError, format, v...) }
