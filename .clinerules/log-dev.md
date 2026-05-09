# Skill: log-dev — Develop and debug the log package

## Usage

When working with the `github.com/kirill-scherba/log` project:

1. Read `docs/*.md` for context (Memory Bank)
2. Check `git status` and GitHub Issues to understand current tasks

## Package structure

```
log/
├── log.go          # Core: Init, Close, public API (Debug/Info/Warn/Error/Sdebug/...)
├── entry.go        # LogEntry struct: String(), Json(), entry()/entryf()
├── loggers.go      # Dispatcher: loggersType, send()
├── es.go           # Elasticsearch output: async bulk sender
├── file.go         # File output: async writer
├── log_test.go     # Tests
├── docs/
│   ├── CONTEXT.md  # Project overview
│   ├── DESIGN.md   # Architecture
│   └── STATUS.md   # Current status
└── .clinerules/
    └── log-dev.md  # This skill
```

## API Reference

### Initialization
- `Init(config *Config)` — initialize loggers, intercept `log.SetOutput()` via `customWriter`
- `Close()` — graceful shutdown via `sync.WaitGroup`

### Typed helpers (send to all active outputs)
- `Debug`, `Debugf`, `Info`, `Infof`, `Warn`, `Warnf`, `Error`, `Errorf`
- `Println`, `Printf` (default level)
- `Fatalln`, `Fatalf`, `Fatal` (Error + os.Exit(1))

### String helpers (return JSON string, do NOT send to outputs)
- `Sdebug`, `Sdebugf`, `Sinfo`, `Sinfof`, `Swarn`, `Swarnf`, `Serror`, `Serrorf`
- `Sentry(level, ...)`, `Sentryf(level, format, ...)`

### Utility
- `SetDefaultLevel(level)` — set default level for standard log calls
- `SetOutput(w io.Writer)` — redirect standard log output

### Log levels
- `LevelDebug` = "DEBUG", `LevelInfo` = "INFO", `LevelWarn` = "WARN", `LevelError` = "ERROR", `LevelNone` = ""

## Config

```go
type Config struct {
    AppShort      string          // Short application name
    AppType       string          // DEV / PROD / TEST
    UseStdout     bool            // Write to stdout
    EsConfig      *EsConfig       // Elasticsearch config (nil = disabled)
    FileConfig    *FileConfig     // File logger config (nil = disabled)
    FilterLevels  []LogLevel      // Levels to suppress
    CustomLogers  []*log.Logger   // Additional loggers to intercept
}
```

## Outputs

### Elasticsearch (es.go)
- Async batch sender
- Config: `URL`, `IndexName`, `EntryChanelSize`, `MaxOverflowBuffer`, `MaxRetryBuffer`
- Uses `http://` (no TLS)

### File (file.go)
- Async writer via goroutine + channel
- Config: `FilePath`
- Writes plain text, not JSON

## Adding a new output

1. Create a file (e.g. `loki.go`)
2. Implement a struct with a `chan *LogEntry`
3. Add a field to `loggersType` in `loggers.go`
4. Add a flag (e.g. `useLokiLogger`)
5. In `send()` add dispatch to the new channel
6. In `Init()` add initialization for the new logger

## Style

- Code and comments in English
- BSD license (see Copyright header in each file)
- Godoc-compatible comments
- Tests in `log_test.go`
- No external dependencies (stdlib only)