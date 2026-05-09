# Status — log

## Current Version

- **Module**: `github.com/kirill-scherba/log`
- **Go Version**: 1.25.7
- **License**: BSD

## What Works

| Feature | Status | Notes |
|---------|--------|-------|
| Stdout logger | ✅ Stable | Console output with timestamp, level, message, fields |
| File logger | ✅ Stable | Async writes via goroutine + channel |
| Elasticsearch logger | ✅ Stable | Async bulk sender with retry, overflow buffer |
| Level filtering | ✅ Stable | `FilterLevels` in `Config` |
| Structured fields | ✅ Stable | Auto-extract `map[string]any` from last argument |
| Standard log compatibility | ✅ Stable | `customWriter` intercepts `log.SetOutput()` |
| Typed helpers (Debug/Info/Warn/Error) | ✅ Stable | Both `Print` and `Printf` variants |
| String helpers (Sdebug/Sinfo/Serror) | ✅ Stable | Return formatted JSON string |
| Graceful shutdown | ✅ Stable | `Close()` via `sync.WaitGroup` |
| Log level parsing from `[LEVEL]` | ✅ Stable | Works with standard `log.Println` |
| Tests | ✅ Stable | Basic test coverage in `log_test.go` |

## Known Issues

- File logger writes plain text, not structured JSON
- No automatic log rotation for file logger
- Elasticsearch config has no TLS/mTLS support (uses `http://`)

## Planned / Future

- [ ] `http.Handler` for runtime log level changes
- [ ] Structured JSON output for file logger
- [ ] Log rotation (size-based) for file logger
- [ ] Custom output writers (syslog, Loki, etc.)
- [ ] OpenTelemetry integration