# Log

[![Go Report Card](https://goreportcard.com/badge/github.com/kirill-scherba/log)](https://goreportcard.com/report/github.com/kirill-scherba/log)
[![GoDoc](https://godoc.org/github.com/kirill-scherba/log?status.svg)](https://godoc.org/github.com/kirill-scherba/log/)

Package log contains log functions to send logs to stdout, files and Elasticsearch.

The log compatible with standard log package, f.e. log.Println are available.

The log messages sends to elasticsearch has json format with basic fields (timestamp, level, message) and additional fields in key-value format.

## Licence

BSD
