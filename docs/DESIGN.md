# Design — log

## Architecture Overview

```
                   ┌─────────────────┐
                   │  log.Println()   │
                   │  log.Printf()    │
                   └────────┬────────┘
                            │
                   ┌────────▼────────┐
                   │  customWriter   │  (implements io.Writer)
                   │  (log.SetOutput)│
                   └────────┬────────┘
                            │
                   ┌────────▼────────┐
                   │  get level from │
                   │  [DEBUG] prefix │
                   └────────┬────────┘
                            │
                   ┌────────▼────────┐
                   │    loggers      │
                   │    .send()      │
                   └───┬────┬────┬───┘
                       │    │    │
           ┌───────────┘    │    └──────────────┐
           ▼                ▼                   ▼
    ┌────────────┐  ┌──────────────┐  ┌──────────────────┐
    │  stdout    │  │  File        │  │  Elasticsearch   │
    │  (console) │  │  (disk)      │  │  (async, batch)  │
    └────────────┘  └──────────────┘  └──────────────────┘
```

## Core Components

### 1. `log.go` — Entry point & public API

- **`Init(config *Config)`** — initializes loggers, sets up `customWriter` as output for standard `log`
- **`Close()`** — gracefully shuts down all loggers, waits for pending writes
- **Typed helpers**: `Debug`, `Info`, `Warn`, `Error` (and their `Printf`/`Println` variants)
- **String helpers**: `Sdebug`, `Sinfo`, `Swarn`, `Serror` (return formatted string, do NOT send to outputs)
- **Utility**: `SetDefaultLevel`, `SetOutput`, `Sentry`, `Sentryf`

### 2. `entry.go` — LogEntry model

- **`LogEntry`** struct:
  - `AppType` — application type (Prod/Dev/Test)
  - `Timestamp` — RFC3339Nano timestamp
  - `Level` — log level string
  - `Message` — log message
  - `Fields` — optional `map[string]any`
- **`String()`** — human-readable format with timestamp, level, message, fields
- **`Json()`** — JSON for Elasticsearch
- **`entry()` / `entryf()`** — constructors with auto-extraction of `Fields` from last argument

### 3. `loggers.go` — Dispatcher

- **`loggersType`** — holds config, flags, and references to ES/file loggers
- **`send(entry)`** — entry point for all outgoing log entries:
  1. Applies `FilterLevels` (skip if level is filtered out)
  2. Writes to stdout (if `useStdoutLogger`)
  3. Pushes to ES channel (if `useEsLogger`)
  4. Pushes to file channel (if `useFailLogger`)

### 4. `file.go` — File logger

- Writes log entries to a file on disk
- Uses a goroutine + channel for async writes
- Configurable file path

### 5. `es.go` — Elasticsearch logger

- Asynchronous bulk sender
- Buffers entries in memory, sends in batches
- Configurable: `entryChanelSize`, `MaxOverflowBuffer`, `MaxRetryBuffer`
- Retry logic with overflow handling

## Key Design Decisions

### `customWriter` interception

The `customWriter` implements `io.Writer` and is set via `log.SetOutput()`. This intercepts all standard `log` calls. It:
1. Strips the standard Go timestamp prefix (`YYYY/MM/DD HH:MM:SS.µs`)
2. Parses log level from `[LEVEL]` at the start of message
3. Falls back to `LevelDefault` (default: `DEBUG`)
4. Sends the resulting `LogEntry` through `loggers.send()`

### Level parsing from standard log

Standard `log.Print` has no concept of levels. The library uses a convention:
- If message starts with `[DEBUG]`, `[INFO]`, `[WARN]`, `[ERROR]` — that level is used
- Otherwise, `LevelDefault` (configurable via `SetDefaultLevel`) is used
- `LevelNone` silences all output

### Graceful shutdown

Both ES and file loggers run in goroutines. `Close()` signals them to stop and waits via `sync.WaitGroup`. This ensures all pending log entries are flushed before the application exits.

### Thread safety

All mutable state is protected by:
- `sync.Mutex` in file logger for write operations
- `sync.WaitGroup` for startup/shutdown coordination
- Channels for goroutine communication (ES, file)

## Filter levels

`Config.FilterLevels` allows suppressing specific log levels entirely. For example, to suppress `DEBUG` in production:
```go
log.Init(&log.Config{
    FilterLevels: []log.LogLevel{log.LevelDebug},
})
```

## Future Considerations

- Add `http.Handler` for runtime log level changes
- Add structured JSON output for file logger (not just plain text)
- Add log rotation for file logger
- Support for custom output writers (e.g., syslog, Loki, etc.)