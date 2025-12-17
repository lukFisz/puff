# AGENTS.md

This file provides comprehensive guidance for AI agents working with the Puff codebase.

## Project Overview

Puff is a Go application that manages Deluge torrent retention via JSON-RPC API. It automatically removes finished torrents after a configurable retention period, with support for cron scheduling, one-time execution, disk space monitoring, and Discord notifications.

**Key characteristics:**
- Small, focused Go project (no tests currently)
- Single binary executable
- Alpine-based Docker images
- External API integration (Deluge JSON-RPC)
- Environment variable configuration

## Essential Commands

### Build and Run

```bash
# Build the application locally
go build -o target/puff

# Run with example environment variables
./run.sh

# Build cross-platform binaries (like CI does)
docker build -f builder.Dockerfile . -t puff-builder
docker run -v "$PWD:/output" --rm puff-builder

# Run tests (currently no tests exist in project)
go test ./...
```

### Docker

```bash
# Build Docker image
docker build -t puff:local .

# Run container with environment variables
docker run --rm \
  -e PUFF_DELUGE_URL="http://deluge.lan/json" \
  -e PUFF_DELUGE_PASSWORD="password" \
  -e PUFF_CRON_SCHEDULE="0 0 * * *" \
  -e PUFF_RETENTION="P14D" \
  puff:local
```

### Development

```bash
# Run with preview mode (mock torrent client)
PUFF_PREVIEW_MODE=true go run main.go

# Download dependencies
go mod download

# Update dependencies
go mod tidy
```

## Project Structure

```
puff/
├── main.go                    # Entry point, initialization, orchestration
├── internal/
│   ├── config.go              # AppConfig struct, env parsing, validation
│   ├── context.go             # AppContext - shared state container
│   ├── job.go                 # Cron scheduling via gocron
│   ├── retentionjob.go        # Main retention logic
│   ├── deluge.go              # DelugeClient - JSON-RPC operations
│   ├── discord.go             # DiscordClient - webhook notifications
│   ├── disk.go                # Disk space checking
│   └── preview.go             # Mock client for preview mode
├── builder.Dockerfile         # Multi-arch build container
├── Dockerfile                 # Runtime container (Alpine)
├── preview.Dockerfile         # Preview gif generation
├── run.sh                     # Local dev script with example env vars
└── .github/workflows/ci.yml   # CI/CD pipeline
```

## Architecture & Flow

### Core Flow

1. **`main.go`**: Entry point
   - Loads and validates config via `GetConfig()` and `ParseValidation()`
   - Initializes logger (multi-writer to stdout and Discord for ERROR/FATAL)
   - Creates `AppContext` with scheduler, clients, config
   - Optional start delay
   - Connection check to Deluge
   - Orchestrates execution with graceful shutdown

2. **`internal/context.go`**: Central state container
   - `AppContext` holds: `Scheduler`, `DiscordClient`, `TorrentClient`, `AppConfig`, `ShutdownChan`
   - Factory function `NewAppContext()` initializes all dependencies
   - Supports preview mode with mock client

3. **`internal/job.go`**: Job scheduling
   - Uses `gocron` library for cron scheduling
   - Supports both recurring cron jobs and one-time execution
   - Event listeners for before/after job execution
   - One-time jobs trigger shutdown via `ShutdownChan`

4. **`internal/retentionjob.go`**: Main business logic
   - `RemoveExpiredTorrents()` is the core retention function
   - Checks disk space threshold (optional)
   - Fetches finished torrents from Deluge
   - Filters by retention period
   - Removes expired torrents with data
   - Logs summary and sends Discord notification

### Key Components

**DelugeClient** (`internal/deluge.go`):
- Interface: `TorrentClient` with methods: `CheckConnection()`, `Login()`, `GetFinishedTorrents()`, `RemoveTorrentsWithData()`
- Handles JSON-RPC authentication (session-based)
- Auto-retries on 401 errors (re-login)
- Timeout configurable via `PUFF_DELUGE_CLIENT_TIMEOUT`
- Fields: `BaseURL`, `Password`, `Session`, `Client`, `idCount`

**DiscordClient** (`internal/discord.go`):
- Implements `io.Writer` to capture ERROR/FATAL logs
- Methods: `SendInfo()` (blue embeds), `SendError()` (red embeds)
- Webhook URL optional - client no-ops if not configured
- Regex check on log output for ERROR/FATAL levels

**AppConfig** (`internal/config.go`):
- Uses `envconfig` for parsing environment variables
- All env vars prefixed with `PUFF_` (except `TZ`)
- Validation methods parse durations, periods, byte sizes
- Helper methods return parsed values (e.g., `RetentionInSeconds()`, `DiskFreeSpaceThresholdInBytes()`)

## Code Conventions

### Package & Module

- Module name: `puff`
- Package structure:
  - `main` package in `main.go`
  - `internal` package for all other code (not intended for external import)

### Imports

- Standard library imports first
- Third-party imports second
- No local package imports (everything is in `internal` or `main`)

### Naming Conventions

- **Struct types**: PascalCase (e.g., `AppConfig`, `DelugeClient`, `TorrentClient`)
- **Interface types**: PascalCase with descriptive names (e.g., `TorrentClient`)
- **Factory functions**: `New` prefix (e.g., `NewAppContext`, `NewDelugeClient`)
- **Public functions**: PascalCase (e.g., `RemoveExpiredTorrents`, `FormattedBytes`)
- **Private functions**: camelCase (e.g., `initLogger`, `delayAppStart`, `sendMessage`)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE (e.g., `appBanner`, `runOnceTag`)
- **Environment variables**: `PUFF_` prefix, SCREAMING_SNAKE_CASE

### Error Handling

- Use `log.Fatal()` for unrecoverable errors during initialization
- Use `log.Error()` for recoverable errors during runtime (with early return)
- Don't return errors from main business logic - log and return
- HTTP client errors are logged but don't crash the application

### Logging

- Use `github.com/charmbracelet/log` library
- Structured logging with key-value pairs: `log.Info("message", "key", value)`
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Sensitive data (passwords) masked with `****` in logs
- Logger configured with true color support

### Configuration

- All configuration via environment variables (no command-line flags)
- Use pointer fields for optional config (e.g., `*string` for `DiskFreeSpaceThreshold`)
- Validation happens early in `ParseValidation()` - fail fast on startup
- Config struct has helper methods that parse and validate values

### Concurrency & Scheduling

- Single job scheduler instance via gocron
- Singleton mode with reschedule policy (jobs don't overlap)
- Graceful shutdown with 1-minute timeout
- Shutdown channel pattern for one-time jobs

## Important Patterns

### JSON-RPC Client Pattern

DelugeClient uses a request/response pattern:
- Auto-incrementing request ID
- Session-based authentication (login once, reuse session)
- Automatic 401 retry with re-login
- Cookie jar for session persistence

### Retry on Authentication Failure

The `request()` method has a `retry401` parameter:
- First call sets `retry401=true`
- On 401 error, calls `Login()` and retries with `retry401=false`
- Prevents infinite retry loops

### Duration/Period Parsing

Multiple formats supported:
- ISO 8601 periods for retention: `P14D`, `P1M`, `P7DT12H`
- Go duration format for timeouts/delays: `2m0s`, `30s`
- Human-readable byte sizes: `100GB`, `1.5TB`, `100GiB`

### Graceful Shutdown

Orchestration pattern in `main.go`:
```go
orchestrateExecution(
    func() *AppContext { return runApp(ctx) },
    gracefulShutdown,
)
```
- Listens for SIGTERM, SIGINT, or shutdown channel
- Calls cleanup function on exit
- Scheduler shutdown with timeout

### Dry Run Mode

Throughout the code:
- Check `ctx.AppConfig.DryRun` before destructive operations
- Log operations that would happen
- Don't skip other logic (still fetch torrents, check retention)

### Preview Mode

Special mode for generating demo GIFs:
- Triggered by `PUFF_PREVIEW_MODE=true` environment variable
- Uses `PreviewTorrentClient` instead of `DelugeClient`
- Returns mock torrents with hardcoded data
- Used in CI for documentation

## Environment Variables

### Required
- `PUFF_DELUGE_URL` - Deluge JSON-RPC endpoint (e.g., `http://deluge.lan/json`)
- `PUFF_DELUGE_PASSWORD` - Deluge password
- `PUFF_CRON_SCHEDULE` - Cron schedule (required unless `PUFF_RUN_ONCE=true`)

### Key Optional
- `PUFF_RUN_ONCE` - Run once and exit (default: `false`)
- `PUFF_DRY_RUN` - Test mode without deletion (default: `false`)
- `PUFF_RETENTION` - ISO 8601 period (default: `P14D`)
- `PUFF_DISK_FREE_SPACE_THRESHOLD` - Minimum free space (e.g., `100GB`)
- `PUFF_DISK_PATH` - Path to monitor (default: `/mnt/puff/monitor`)
- `PUFF_DISCORD_WEBHOOK_URL` - Discord notifications
- `PUFF_LOG_LEVEL` - Logging level (default: `INFO`)
- `PUFF_START_DELAY` - Startup delay (default: `0s`)
- `PUFF_DELUGE_CLIENT_TIMEOUT` - HTTP timeout (default: `2m0s`)
- `TZ` - Timezone (default: `UTC`)

## Dependencies

Key third-party libraries:
- `github.com/go-co-op/gocron/v2` - Cron job scheduling
- `github.com/kelseyhightower/envconfig` - Environment variable parsing
- `github.com/charmbracelet/log` - Structured logging
- `github.com/rickb777/date` - ISO 8601 period parsing
- `github.com/dustin/go-humanize` - Human-readable byte sizes
- `github.com/google/uuid` - UUID generation (for job IDs)

## CI/CD Pipeline

### Trigger
- On git tags matching `v*` pattern

### Process
1. Extract version from tag (and short version)
2. Detect RC tags (contain "rc" or "RC")
3. Build multi-arch binaries (amd64, arm64) using `builder.Dockerfile`
4. Generate preview.gif (non-RC only) and push to master
5. Build and push Docker images to GitHub Container Registry
6. Tag images:
   - RC: only exact version tag
   - Release: `latest`, short version (e.g., `v0.4`), exact version
7. Generate changelog from git commits
8. Create GitHub release (prerelease for RC tags)

### Docker Images
- Published to `ghcr.io/lukfisz/puff`
- Multi-platform support: `linux/amd64`, `linux/arm64`
- Based on Alpine 3.22 (small footprint)
- Includes tzdata for timezone support

## Testing

### Current State
- **No tests currently exist in the project**
- `go test ./...` will run but find no tests
- This is documented in CLAUDE.md

### If Adding Tests
- Place test files adjacent to source: `internal/config_test.go`
- Use standard Go testing: `package internal` with `_test.go` suffix
- Mock `TorrentClient` interface for unit tests (see `PreviewTorrentClient` as example)
- Test helper methods on `AppConfig` (parsing, validation)
- Test `RemoveExpiredTorrents` logic with mock client

## Gotchas & Non-Obvious Patterns

### 1. Config Validation is Eager
All config parsing happens in `ParseValidation()` at startup. This includes:
- Parsing durations, periods, byte sizes
- Loading timezone
- Checking disk path accessibility
If any validation fails, the app exits immediately with `log.Fatal()`.

### 2. Logger is Multi-Writer
The logger writes to both stdout and Discord:
```go
multi := io.MultiWriter(os.Stdout, puff.NewDiscordClient(config))
log.SetOutput(multi)
```
DiscordClient implements `io.Writer` and filters for ERROR/FATAL levels.

### 3. Session Persistence in DelugeClient
The Deluge session cookie is stored in the HTTP client's cookie jar, not as a separate field. The `Session` field in `DelugeClient` is for tracking but the actual session comes from cookies.

### 4. One-Time Jobs Trigger Shutdown
When `PUFF_RUN_ONCE=true`, the job is tagged with `runOnceTag`. After execution, the `afterJob` listener detects this tag and sends to `ShutdownChan`, triggering graceful shutdown.

### 5. Disk Threshold is Optional Gatekeeper
If `PUFF_DISK_FREE_SPACE_THRESHOLD` is set:
- Job checks disk space before retention logic
- If disk has enough space, retention is skipped entirely
- If threshold exceeded, retention runs normally

### 6. DryRun Only Affects Deletion
`PUFF_DRY_RUN=true` prevents actual torrent removal but:
- Still connects to Deluge
- Still fetches torrents
- Still logs what would be removed
- Still calculates space savings

### 7. ISO 8601 Period Approximation
Retention uses `period.DurationApprox()` which approximates months/years:
- 1 month ≈ 30 days
- 1 year ≈ 365 days
For precise retention, use days (e.g., `P30D` instead of `P1M`).

### 8. Cron Format is 5 or 6 Fields
Uses `robfig/cron/v3` syntax:
- 5 fields: minute, hour, day, month, weekday
- 6 fields: second, minute, hour, day, month, weekday
- Special formats: `@every 5s`, `@daily`, etc.

### 9. No Makefile
The project doesn't use Make. All build commands are either:
- Direct `go build` commands
- Docker builds
- Shell scripts (`run.sh`)

### 10. Version Injection via Build Args
Version is injected at Docker build time via `PUFF_CURRENT_VERSION` build arg, which becomes an environment variable read by the app.

## Common Tasks

### Add a New Environment Variable
1. Add field to `AppConfig` struct in `internal/config.go`
2. Add `envconfig` tag with `PUFF_` prefix
3. If requires parsing, add validation method (e.g., `MyConfigDuration()`)
4. Call validation method in `ParseValidation()`
5. Update `logConfig()` in `main.go` to include in startup logs
6. Update README.md and this file

### Add a New Torrent Client Operation
1. Add method to `TorrentClient` interface in `internal/deluge.go`
2. Implement in `DelugeClient` struct
3. Implement stub in `PreviewTorrentClient` (for preview mode)
4. Use JSON-RPC `request()` helper
5. Handle authentication with `retry401` parameter

### Modify Retention Logic
1. Edit `RemoveExpiredTorrents()` in `internal/retentionjob.go`
2. Keep disk threshold check at the top
3. Respect `DryRun` mode for destructive operations
4. Update Discord notification if output changes
5. Update logs for visibility

### Add a New Job Type
1. Create job function (signature: `func()`)
2. Call either `ScheduleCronjob()` or `NewOneTimeJob()` in `runApp()`
3. Job will automatically get before/after listeners
4. For one-time, add `runOnceTag` to trigger shutdown

### Modify Docker Build
- **Runtime changes**: Edit `Dockerfile`
- **Build changes**: Edit `builder.Dockerfile`
- **CI changes**: Edit `.github/workflows/ci.yml`
- Multi-arch is handled by Docker Buildx in CI

## Resources

- [Deluge JSON-RPC API Reference](https://deluge.readthedocs.io/en/latest/reference/api.html)
- [gocron Documentation](https://github.com/go-co-op/gocron)
- [ISO 8601 Duration Format](https://en.wikipedia.org/wiki/ISO_8601#Durations)
- [Cron Expression Format](https://pkg.go.dev/github.com/robfig/cron/v3)

## Questions to Ask When Working on This Codebase

- Does this change require environment variable configuration?
- Should this be logged? At what level?
- Does this respect dry-run mode?
- Is this operation idempotent (safe to retry)?
- Will this work in preview mode?
- Does this need Discord notification?
- Is error handling appropriate (fatal vs recoverable)?
- Are we logging sensitive data?
- Does this affect the retention logic?
- Should this be documented in README.md?
