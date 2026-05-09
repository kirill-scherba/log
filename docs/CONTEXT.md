# Context — log

## Project Overview

Package `log` provides a flexible logging library for Go applications. It sends log messages simultaneously to multiple outputs: **stdout**, **file**, and **Elasticsearch**.

The library is compatible with the standard `log` package — calls like `log.Println()` are intercepted and routed through the same pipeline.

## Key Features

- **Multi-output**: stdout (console), file (disk), Elasticsearch (JSON)
- **Standard log compatibility**: `log.SetOutput()` is intercepted via `customWriter`
- **Log levels**: `DEBUG`, `INFO`, `WARN`, `ERROR`, `NONE`
- **Structured fields**: last argument `map[string]any` auto-extracted as `Fields`
- **Level filtering**: `FilterLevels` in `Config` to suppress specific levels
- **Elasticsearch support**: asynchronous bulk sending with retry and overflow buffer
- **File logger**: write to disk with configurable path
- **JSON output**: `LogEntry.Json()` for ES-compatible JSON format
- **Graceful shutdown**: `Close()` waits for all loggers via `sync.WaitGroup`
- **Type-safe helpers**: `Sdebug`, `Sinfo`, `Swarn`, `Serror` (return string) and `PrintLevel`, `Debug`, `Info`, `Warn`, `Error` (send to outputs)

## Target Audience

- Go developers who need structured logging with multiple backends
- Projects that want both console and Elasticsearch logging
- Developers familiar with standard `log` looking for a drop-in replacement with more power

## Dependencies

- **Go 1.25+** — build and runtime
- **No external dependencies** — pure standard library

## Related Projects

- [memory-store-mcp](https://github.com/kirill-scherba/memory-store-mcp) — MCP server that uses the same logging patterns
- [db-tool-mcp](https://github.com/kirill-scherba/db-tool-mcp) — another project with bot logging

## License

BSD