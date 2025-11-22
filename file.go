package log

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// FileConfig is a struct that holds information about how to send log entries
// to a file.
type FileConfig struct {
	// Log files folder
	Folder string

	// Create new log file after
	CreateNewAfter time.Duration
}

// file is a struct that holds information about how to send log entries to a
// file.
type file struct {

	// fileEntryChannel is a channel that receives log entries for sending to
	// file
	fileEntryChannel chan *LogEntry

	// File log parameters
	*FileConfig

	// Application short name
	AppShort string

	// Current opened log file
	f *os.File

	// File log created time
	fCreatedAt time.Time
}

// init sets up the file logger and starts the entry handler goroutine.
func (f *file) init(appShort string, fileConfig *FileConfig) {
	if fileConfig == nil {
		return
	}

	// Set file logger config
	f.FileConfig = fileConfig
	f.AppShort = appShort

	// Create entry channel
	f.fileEntryChannel = make(chan *LogEntry, 100)

	// Start entry handler
	loggers.wgStart.Add(1)
	go f.entryHandler()
}

// close closes the entry channel and stop the entry processing goroutine.
func (f *file) close() {
	close(f.fileEntryChannel)
}

// entryHandler is a goroutine that consumes log entries from the fileEntryChannel.
// It checks if the log entry channel is closed, and if so, it exits the goroutine.
// It then either creates a new file, or switches to a new file after a certain
// time period. Finally, it sends the log entries to file.
func (f *file) entryHandler() {
	loggers.wgStart.Done()

	loggers.wgClose.Add(1)
	defer loggers.wgClose.Done()

	// Loop until the goroutine is stopped
	for {
		entry, ok := <-f.fileEntryChannel
		if !ok {
			// If the channel is closed, exit the goroutine
			break
		}

		// Set or change file
		var err error

		switch f.f {

		// Create new file
		case nil:
			err = f.newLogfile()

		// Switch file
		default:
			// If file log created more than 10 minutes ago
			if f.CreateNewAfter > 0 && time.Since(f.fCreatedAt) > f.CreateNewAfter {
				// Close current file
				f.f.Close()

				// Create new file
				err = f.newLogfile()
			}
		}
		if err != nil {
			continue
		}

		// Send to file
		f.f.Write([]byte(entry.String() + "\n"))
	}
}

// newLogfile creates a new log file and switches the file logger to it.
// It creates a new folder if the folder does not exist and creates a new log
// file with the format "appshort_timestamp.log". It then compresses the old
// log file after 1 second and removes the old log file.
func (f *file) newLogfile() (err error) {
	var now = time.Now()
	timeStr := now.Format("2006.01.02-15.04.05")

	folder := f.FileConfig.Folder
	if folder == "" {
		folder = os.TempDir()
	}
	folder = folder + "/" + f.AppShort

	// Create folder if not exists
	if _, err = os.Stat(folder); os.IsNotExist(err) {
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			return
		}
	}

	// Create new log file
	fileName := fmt.Sprintf("%s/%s_%s.log", folder, f.AppShort, timeStr)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("error creating log file:", err)
		return
	}

	// Compress end remove old file after 1 second
	if f.f != nil {
		fileName := f.f.Name()
		time.AfterFunc(1*time.Second, func() {
			time.Sleep(1 * time.Second)
			f.compressFile(fileName)
			os.Remove(fileName)
		})
	}

	// Set new file
	f.f = file
	f.fCreatedAt = now
	log.Println("create new log file:", file.Name())
	return
}

// compressFile compresses the log file given by name.
func (f *file) compressFile(name string) (err error) {

	// Open srcFile
	srcFile, err := os.Open(name)
	if err != nil {
		return
	}
	defer srcFile.Close()

	// Create new file
	dstFile, err := os.Create(name + ".gz")
	if err != nil {
		return
	}
	defer dstFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(dstFile)
	defer gzipWriter.Close()

	// Copy data from srcFile to gzipWriter
	if _, err = io.Copy(gzipWriter, srcFile); err != nil {
		err = fmt.Errorf("error compressing log file: %w", err)
		return
	}

	return
}
