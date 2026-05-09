# Log

[![Go Report Card](https://goreportcard.com/badge/github.com/kirill-scherba/log)](https://goreportcard.com/report/github.com/kirill-scherba/log)
[![GoDoc](https://godoc.org/github.com/kirill-scherba/log?status.svg)](https://godoc.org/github.com/kirill-scherba/log/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kirill-scherba/log)](https://golang.org)

Package `log` provides a flexible logging library for Go applications. It sends log messages simultaneously to **stdout**, **file**, and **Elasticsearch** — with structured fields, level filtering, and graceful shutdown.

Compatible with the standard `log` package — calls like `log.Println()` are intercepted and routed through the same pipeline.

---

## Features

- **Multi-output**: stdout (console), file (disk), Elasticsearch (JSON)
- **Standard log compatibility**: `log.SetOutput()` interception via `customWriter`
- **Log levels**: `DEBUG`, `INFO`, `WARN`, `ERROR`, `NONE`
- **Structured fields**: last argument `map[string]any` auto-extracted as `Fields`
- **Level filtering**: suppress specific levels via `Config.FilterLevels`
- **Elasticsearch**: asynchronous bulk sending with retry and overflow buffer
- **File logger**: write to disk with configurable path
- **JSON output**: `LogEntry.Json()` for ES-compatible format
- **Graceful shutdown**: `Close()` waits for all loggers via `sync.WaitGroup`
- **Type-safe helpers**: `Debug`, `Info`, `Warn`, `Error` + string variants `Sdebug`, `Sinfo`, etc.

---

## Installation

```bash
go get github.com/kirill-scherba/log
```

---

## Quick Start

```go
package main

import (
    "log"
    "github.com/kirill-scherba/log"
)

func main() {
    // Initialize with stdout only
    log.Init(&log.Config{
        AppShort: "myapp",
        AppType:  "DEV",
        UseStdout: true,
    })
    defer log.Close()

    // Use compatible standard log calls
    log.Println("Hello from standard log")

    // Use typed helpers with levels
    log.Debug("debug message")
    log.Info("info message")
    log.Warn("warn message")
    log.Error("error message")

    // With structured fields (auto-detected map[string]any)
    log.Info("user action", map[string]any{"user": "john", "action": "login"})

    // Formatted output with fields
    log.Infof("processed %d records", 42, map[string]any{"batch": "nightly"})
}
```

Output:
```
2025/10/19 10:28:50.567024 [DEBUG] Hello from standard log
2025/10/19 10:28:50.567025 [DEBUG] debug message
2025/10/19 10:28:50.567026 [INFO]  info message
2025/10/19 10:28:50.567027 [WARN]  warn message
2025/10/19 10:28:50.567028 [ERROR] error message
2025/10/19 10:28:50.567029 [INFO]  user action, fields: map[user:john action:login]
```

---

## Configuration

| Config Field     | Type           | Description                                          |
|------------------|----------------|------------------------------------------------------|
| `AppShort`       | `string`       | Short application name                               |
| `AppType`        | `string`       | Application type: `DEV`, `PROD`, `TEST`              |
| `UseStdout`      | `bool`         | Write to stdout (default: `true` on Init)            |
| `EsConfig`       | `*EsConfig`    | Elasticsearch config (nil = disabled)                |
| `FileConfig`     | `*FileConfig`  | File logger config (nil = disabled)                  |
| `FilterLevels`   | `[]LogLevel`   | Levels to suppress                                   |
| `CustomLogers`   | `[]*log.Logger`| Additional loggers to intercept                      |

### Elasticsearch

```go
log.Init(&log.Config{
    AppShort: "myapp",
    AppType:  "PROD",
    UseStdout: true,
    EsConfig: &log.EsConfig{
        URL:           "http://localhost:9200",
        IndexName:     "myapp-logs",
        EntryChanelSize: 100,
    },
})
```

### File

```go
log.Init(&log.Config{
    AppShort: "myapp",
    AppType:  "PROD",
    UseStdout: true,
    FileConfig: &log.FileConfig{
        FilePath: "/var/log/myapp/app.log",
    },
})
```

### Level filtering (suppress DEBUG in production)

```go
log.Init(&log.Config{
    AppShort:     "myapp",
    AppType:      "PROD",
    UseStdout:    true,
    FilterLevels: []log.LogLevel{log.LevelDebug},
})
```

---

## API Reference

### Initialization

| Function | Description |
|----------|-------------|
| `Init(config *Config)` | Initialize loggers and intercept standard `log` |
| `Close()` | Flush and shut down all loggers |

### Output helpers (send to all active loggers)

| Function | Level | Description |
|----------|-------|-------------|
| `Debug(v ...any)` | DEBUG | Log with debug level |
| `Debugf(format, v ...any)` | DEBUG | Formatted debug log |
| `Info(v ...any)` | INFO | Log with info level |
| `Infof(format, v ...any)` | INFO | Formatted info log |
| `Warn(v ...any)` | WARN | Log with warn level |
| `Warnf(format, v ...any)` | WARN | Formatted warn log |
| `Error(v ...any)` | ERROR | Log with error level |
| `Errorf(format, v ...any)` | ERROR | Formatted error log |
| `Println(v ...any)` | default | Debug-level log (standard compatibility) |
| `Printf(format, v ...any)` | default | Formatted debug log |
| `Fatalln(v ...any)` | ERROR | Error + `os.Exit(1)` |
| `Fatalf(format, v ...any)` | ERROR | Formatted error + `os.Exit(1)` |
| `Fatal(v ...any)` | ERROR | Error + `os.Exit(1)` |

### String helpers (return JSON string, do NOT send to outputs)

| Function | Level | Description |
|----------|-------|-------------|
| `Sdebug(v ...any)` | DEBUG | Return debug JSON string |
| `Sdebugf(format, v ...any)` | DEBUG | Return formatted debug JSON string |
| `Sinfo(v ...any)` | INFO | Return info JSON string |
| `Sinfof(format, v ...any)` | INFO | Return formatted info JSON string |
| `Swarn(v ...any)` | WARN | Return warn JSON string |
| `Swarnf(format, v ...any)` | WARN | Return formatted warn JSON string |
| `Serror(v ...any)` | ERROR | Return error JSON string |
| `Serrorf(format, v ...any)` | ERROR | Return formatted error JSON string |
| `Sentry(level, v ...any)` | any | Return JSON string at specified level |
| `Sentryf(level, format, v ...any)` | any | Return formatted JSON string at specified level |

### Utility

| Function | Description |
|----------|-------------|
| `SetDefaultLevel(level)` | Set default level for standard log calls |
| `SetOutput(w io.Writer)` | Redirect standard log output |

---

## Log Levels

| Constant   | Value   | Description                |
|------------|---------|----------------------------|
| `LevelDebug` | `DEBUG` | Detailed debug information |
| `LevelInfo`  | `INFO`  | Informational messages     |
| `LevelWarn`  | `WARN`  | Warning messages           |
| `LevelError` | `ERROR` | Error messages             |
| `LevelNone`  | `""`    | Silence all standard logs  |

---

## How It Works

```
   log.Println() ──► customWriter ──► parse [LEVEL] ──► loggers.send()
                                                           ├── stdout
                                                           ├── file (async)
                                                           └── Elasticsearch (async, batch)
```

The `customWriter` implements `io.Writer` and is set via `log.SetOutput()`. It:
1. Strips standard Go timestamp
2. Parses log level from `[LEVEL]` prefix (falls back to `LevelDefault`)
3. Sends the resulting `LogEntry` through the dispatch pipeline

---

## License

BSD