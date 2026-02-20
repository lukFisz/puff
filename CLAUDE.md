# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run

```bash
# Build the application
go build -o target/puff

# Run locally (uses run.sh with example env vars)
./run.sh

# Run tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run tests verbosely
go test ./... -v
```

## Environment Variables

Required (at least one torrent client must be configured):
- `PUFF_DELUGE_URL` - Deluge JSON-RPC API endpoint (e.g., `http://deluge.lan/json`)
- `PUFF_DELUGE_PASSWORD` - Deluge password
- `PUFF_QBITTORRENT_URL` - qBittorrent Web API endpoint (e.g., `http://qbit.lan:8080`)
- `PUFF_QBITTORRENT_USERNAME` - qBittorrent username
- `PUFF_QBITTORRENT_PASSWORD` - qBittorrent password
- `PUFF_CRON_SCHEDULE` - Cron schedule (required unless `PUFF_RUN_ONCE=true`)

Key optional:
- `PUFF_RUN_ONCE` - Run once and exit instead of scheduling (default: `false`)
- `PUFF_DRY_RUN` - Test mode without actual deletion (default: `false`)
- `PUFF_RETENTION` - ISO 8601 period format (default: `P14D`)
- `PUFF_TORRENT_CLIENT_TIMEOUT` - HTTP client timeout (default: `2m0s`)
- `PUFF_DISK_FREE_SPACE_THRESHOLD` - Skip retention if disk has enough space (e.g., `100GB`)
- `PUFF_DISCORD_WEBHOOK_URL` - Discord notifications for job summaries/errors

## Architecture

Puff is a Go application that manages torrent retention for Deluge and qBittorrent clients.

### Core Flow
1. `main.go` - Entry point, initializes config/logging, orchestrates execution with graceful shutdown
2. `internal/context.go` - `AppContext` holds shared state (scheduler, jobs list, config, shutdown channel). `Jobs` is a list of `TorrentClientJob` (one per configured client).
3. `internal/job.go` - Cron scheduling via gocron, supports one-time and recurring jobs
4. `internal/retentionjob.go` - Main retention logic: checks disk threshold, filters expired torrents, removes them. Called per torrent client.

### Interfaces
- `internal/torrent.go` - Defines `TorrentClient` interface (CheckConnection, Login, GetFinishedTorrents, RemoveTorrentsWithData) and `Torrent` interface (Expired, Id, SizeInBytes, Name)

### Clients
- `internal/deluge.go` - `DelugeClient` implements `TorrentClient` via JSON-RPC. `TorrentDeluge` implements `Torrent`.
- `internal/qbittorrent.go` - `QbitTorrentClient` implements `TorrentClient` via qBittorrent Web API (resty). `TorrentQbit` implements `Torrent`.
- `internal/discord.go` - `DiscordClient` sends webhook notifications; also implements `io.Writer` to capture ERROR/FATAL logs

### Supporting
- `internal/config.go` - `AppConfig` struct with envconfig parsing and validation methods
- `internal/disk.go` - Disk free space checking via syscall
- `internal/utils.go` - Utility functions (FormattedBytes)
- `internal/preview.go` - `PreviewTorrentClient` for preview/demo mode
