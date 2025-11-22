package log

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EsConfig is a struct that holds information about how to send log entries to
// Elasticsearch.
//
// Elasticsearch index creation in Kibana Dev Tools:
//
// Delete existing index:
/*
DELETE /app-name-index
*/
// Create new index:
/*
PUT /app-name-index
{
  "mappings": {
    "properties": {
      "@timestamp": {
        "type": "date_nanos"
      },
      "app_type": {
        "type": "keyword"
      },
      "level": {
        "type": "keyword"
      },
      "message": {
        "type": "text"
      },
      "fields": {
        "type": "object"
      }
    }
  }
}
*/
type EsConfig struct {
	ES_URL        string // Elasticsearch URL
	ES_API_KEY    string // Elasticsearch API key
	ES_INDEX_NAME string // Elasticsearch index name

	// Time to hold before sending log entries to Elasticsearch.
	// If not set, Default is 5 seconds.
	TimeToHold time.Duration

	// Maximum number of log entries to hold before sending them to Elasticsearch.
	// If not set, Default is 1000.
	EntriesToHold int

	// Directory to store failed batches on disk.
	// If not set, Default is "/tmp/APP_SHORT_NAME/failover".
	FailoverDir string

	// Maximum number of failover files to keep on disk.
	// If not set, Default is 10.
	MaxFailoverFiles int
}

// es is a struct that holds information about how to send log entries to
// Elasticsearch.
type es struct {

	// esEntryChannel is a channel that receives log entries for sending to es
	esEntryChannel chan *LogEntry

	// Elasticsearch log parameters
	*EsConfig
}

// init sets up the Elasticsearch logger and starts the entry handler goroutine.
//
// The entry handler goroutine aggregates log entries in a slice until either
// the slice reaches the maximum size (l.entriesToHold) or the time to hold
// (l.timeToHold) expires.
// When either condition is met, it sends the aggregated log entries to
// Elasticsearch using the sendToElasticsearch method.
// If the Elasticsearch config is not set, the function does nothing.
// It also sets the default values for the Elasticsearch config if they are not
// set.
func (e *es) init(appShort string, esConfig *EsConfig) {
	if esConfig == nil {
		return
	}

	// Set elasticsearch logger config
	e.EsConfig = esConfig

	// Set failover directory
	if e.EsConfig.FailoverDir == "" {
		tempDir := os.TempDir()
		e.EsConfig.FailoverDir = tempDir + "/" + appShort + "/failover"
	}
	os.MkdirAll(e.EsConfig.FailoverDir, 0755)

	// Set default time to hold
	if e.EsConfig.TimeToHold == 0 {
		e.EsConfig.TimeToHold = 10 * time.Second
	}

	// Set default entries to hold
	if e.EsConfig.EntriesToHold == 0 {
		e.EsConfig.EntriesToHold = 1000
	}

	// Set default max failover files
	if e.EsConfig.MaxFailoverFiles == 0 {
		e.EsConfig.MaxFailoverFiles = 10
	}

	// Create entry channel
	e.esEntryChannel = make(chan *LogEntry, esConfig.EntriesToHold)

	// Start entry handler
	loggers.wgStart.Add(1)
	go e.entryHandler()
}

// close closes the entry channel and stop the entry processing goroutine.
func (e *es) close() {
	close(e.esEntryChannel)
}

// entryHandler is a goroutine that consumes log entries from the entryChannel.
// It aggregates log entries in a slice until either the slice reaches the maximum
// size (l.entriesToHold) or the time to hold (l.timeToHold) expires.
// When either condition is met, it sends the aggregated log entries to Elasticsearch
// using the sendToElasticsearch method.
// If sending fails, it buffers the batch for later retries.
func (e *es) entryHandler() {
	loggers.wgStart.Done()

	loggers.wgClose.Add(1)
	defer loggers.wgClose.Done()

	// Slice to hold log entries
	var entries []*LogEntry
	ticker := time.NewTicker(e.TimeToHold)
	defer ticker.Stop()

	// Loop until the goroutine is stopped
	for {
		// First, try to send any buffered batches
		if e.processFailoverFiles() {
			// If we successfully sent a file, continue the loop to immediately try the next one.
			continue
		}

		select {

		// Case for when a new log entry is received
		case entry, ok := <-e.esEntryChannel:
			if !ok {
				// Before exiting, try to send any remaining entries
				if len(entries) > 0 {
					e.sendOrSave(entries)
				}
				// If the channel is closed, exit the goroutine
				return
			}
			// Append the new log entry to the slice
			entries = append(entries, entry)

			// If the slice has reached the maximum size, send the log entries
			if len(entries) >= e.EntriesToHold {
				e.sendOrSave(entries)
				entries = nil
				ticker.Reset(e.TimeToHold) // Reset ticker after sending a full batch
			}

		// Case for when the time to hold expires
		case <-ticker.C:
			// If there are any log entries in the slice, send them to Elasticsearch
			if len(entries) > 0 {
				e.sendOrSave(entries)
				entries = nil
			}
		}
	}
}

// sendOrSave attempts to send a batch of entries, and if it fails, saves it
// to a failover file on disk.
func (e *es) sendOrSave(entries []*LogEntry) {
	err := e.sendToElasticsearch(entries...)
	if err != nil {
		stdoutLogger.Println(
			"error sending log entries to Elasticsearch, saving to disk for retry:",
			err)

		// On failure, save the batch to a disk file.
		if err := e.saveBatchToDisk(entries); err == nil {
			stdoutLogger.Println("successfully saved failed batch to disk")
		} else {
			stdoutLogger.Println("CRITICAL: Failed to save batch to disk:", err)
		}
	}
}

// saveBatchToDisk saves a slice of LogEntry to a unique file in the failover directory.
func (e *es) saveBatchToDisk(entries []*LogEntry) error {
	if e.FailoverDir == "" {
		return fmt.Errorf("FailoverDir is not configured")
	}

	// Check and enforce MaxFailoverFiles limit
	files, err := os.ReadDir(e.FailoverDir)
	if err != nil {
		return fmt.Errorf("could not read failover directory: %w", err)
	}
	if len(files) >= e.MaxFailoverFiles {
		return fmt.Errorf(
			"max failover files limit (%d) reached, discarding batch",
			e.MaxFailoverFiles)
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal batch for disk save: %w", err)
	}

	fileName := fmt.Sprintf("batch-%d.json", time.Now().UnixNano())
	filePath := filepath.Join(e.FailoverDir, fileName)

	return os.WriteFile(filePath, data, 0644)
}

// processFailoverFiles checks for and processes one file from the failover directory.
// It returns true if a file was successfully processed and deleted, false otherwise.
func (e *es) processFailoverFiles() bool {
	if e.FailoverDir == "" {
		return false
	}

	// Check for failover files
	files, err := filepath.Glob(filepath.Join(e.FailoverDir, "*.json"))
	if err != nil || len(files) == 0 {
		return false
	}

	// Sort files by modification time
	sort.Strings(files)
	filePath := files[0]

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		stdoutLogger.Printf("error reading failover file %s: %v", filePath, err)
		return false
	}

	// Unmarshal the data
	var entries []*LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		stdoutLogger.Printf("error unmarshalling failover file %s: %v, deleting corrupt file.", filePath, err)
		os.Remove(filePath)
		return false
	}

	// Attempt to send the batch
	stdoutLogger.Printf("attempting to send batch from failover file: %s", filePath)
	if err := e.sendToElasticsearch(entries...); err == nil {
		stdoutLogger.Printf("successfully sent batch from %s, deleting file.", filePath)
		os.Remove(filePath)
		return true
	}

	// If sending fails, retry later
	stdoutLogger.Printf("failed to send batch from %s, will retry later: %v", filePath, err)
	return false
}

// sendToElasticsearch Sends entrys to Elasticsearch
func (e *es) sendToElasticsearch(entrys ...*LogEntry) (err error) {

	// Check that Elasticsearch config is set
	if e.EsConfig == nil {
		err = fmt.Errorf("elasticsearch config is not set")
		return
	}

	// Create string buffer and reader
	var buf strings.Builder
	for _, entry := range entrys {
		json := string(entry.Json())
		buf.WriteString(fmt.Sprintf(`{ "index": { "_index": "%s" } }`+"\n", e.ES_INDEX_NAME))
		buf.WriteString(json + "\n")
	}

	// Create a gzip writer
	var gzipBuf bytes.Buffer
	gz := gzip.NewWriter(&gzipBuf)

	// Write the buffered string into the gzip writer and close it
	if _, err := gz.Write([]byte(buf.String())); err != nil {
		return fmt.Errorf("Error writing to gzip writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("Error closing gzip writer: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", e.ES_URL+"/_bulk?pretty&pipeline=ent-search-generic-ingestion", &gzipBuf)
	if err != nil {
		err = fmt.Errorf("Error creating HTTP request: %v", err)
		return
	}
	req.Header.Set("Authorization", "ApiKey "+e.ES_API_KEY)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Execute HTTP request with 10 second timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("Error sending HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check response status and body if error
	if resp.StatusCode != http.StatusOK {

		// Get response status
		responseStatus := fmt.Sprintf("Error Response Status: %s", resp.Status)

		// Get response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("%s\nError reading response body: %v", responseStatus, err)
			return err
		}
		responseBody := fmt.Sprintf("Response Body: %s", string(body))

		// Return error
		err = fmt.Errorf("%s\n%s", responseStatus, responseBody)
		return err
	}

	return
}
