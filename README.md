# <img src="/img/logo.png" alt="drawing" width="70" style="transform: translate(0, 15px)"/> Puff - Torrent Retention Manager

Application for automatic torrent management. Removes finished torrents that have exceeded a specified
retention period. Supports Deluge (JSON-RPC) and qBittorrent (Web API).


![preview.gif](img/preview.gif)

## Description

Puff is a background application that connects to torrent clients (Deluge via JSON-RPC, qBittorrent via Web API)
and automatically removes finished torrents according to a configurable retention policy.
The application can run as a cron job scheduler or execute the retention job once and exit.
It can be deployed as a Docker container with a small footprint. Multiple torrent clients can be configured simultaneously.

## Features

- 🔄 Automatic removal of finished torrents after retention period
- 🧲 Support for Deluge and qBittorrent clients
- ⏰ Cron-based job scheduling
- 🏃‍♂️ Run-once execution
- 🧪 Dry-run mode for testing without actual deletion
- 🔐 Authentication via Deluge JSON-RPC API and qBittorrent Web API
- 📊 Detailed operation logging
- 🐳 Ready-to-use Docker image
- 💬 Discord notifications for job summaries and errors
- 💾 Disk space monitoring

## Requirements

- Go 1.25
- Access to a Deluge instance (WebUI + JSON-RPC API) and/or qBittorrent instance (Web API)
- (Optional) Docker for containerized deployment

## Configuration

The application is configured via environment variables:

| Variable                         | Description                                                         | Required | Default Value |
|----------------------------------|---------------------------------------------------------------------|----------|---------------|
| `PUFF_CRON_SCHEDULE`             | Cron schedule (e.g., `0 0 * * *` for daily at midnight). Required if `PUFF_RUN_ONCE` is `false`. | No       | -             |
| `PUFF_DELUGE_URL`                | URL to Deluge JSON-RPC API (e.g., `http://deluge.lan/json`)         | No       | -             |
| `PUFF_DELUGE_PASSWORD`           | Deluge password                                                     | No       | -             |
| `PUFF_QBITTORRENT_URL`           | URL to qBittorrent Web API (e.g., `http://qbit.lan:8080`)          | No       | -             |
| `PUFF_QBITTORRENT_USERNAME`      | qBittorrent username                                                | No       | -             |
| `PUFF_QBITTORRENT_PASSWORD`      | qBittorrent password                                                | No       | -             |
| `PUFF_RUN_ONCE`                  | If `true`, the application runs the retention job once and exits.   | No       | `false`       |
| `PUFF_TORRENT_CLIENT_TIMEOUT`    | HTTP client timeout. Valid time units: ns, us (or µs), ms, s, m, h. | No       | `2m0s`        |
| `PUFF_RETENTION`                 | Retention period in ISO 8601 format (e.g., `P14D` = 14 days)       | No       | `P14D`        |
| `PUFF_START_DELAY`               | Application startup delay                                           | No       | `0s`          |
| `PUFF_DRY_RUN`                   | Test mode (does not delete torrents)                                | No       | `false`       |
| `PUFF_LOG_LEVEL`                 | Logging level (DEBUG, INFO, WARN, ERROR)                            | No       | `INFO`        |
| `PUFF_DISK_FREE_SPACE_THRESHOLD` | Minimum free disk space threshold (e.g., `100GB`, `1.5TB`, `100GiB`). If set, torrents will be removed when free space falls below this value. | No       | -             |
| `PUFF_DISK_PATH`                 | Path to the disk to monitor for free space                          | No       | `/mnt/puff/monitor` |
| `PUFF_DISCORD_WEBHOOK_URL`       | Discord webhook URL for notifications                               | No       | -             |
| `TZ`                             | Time zone                                                           | No       | `UTC`         |

At least one torrent client must be configured (`PUFF_DELUGE_URL` and/or `PUFF_QBITTORRENT_URL`).

### Retention Format

Retention is specified in ISO 8601 period format:

- `P14D` - 14 days
- `P30D` - 30 days
- `P1M` - 1 month
- `P7DT12H` - 7 days and 12 hours

### Cron Schedule Format

Uses standard cron format with 5 or 6 fields:

- `@every 5s` - every 5 seconds
- `0 0 * * *` - daily at midnight
- `0 */6 * * *` - every 6 hours
- `0 0 * * 0` - weekly on Sunday
- `0/10 * * * * *` - every 10 seconds (6 fields, with seconds)

## Usage

### Local Execution

Use [run file](./run.sh)

### Docker Execution

Run with Deluge:
```bash
docker run -d \
  -e PUFF_CRON_SCHEDULE="0 0 * * *" \
  -e PUFF_DELUGE_URL="http://url_or_ip_to_deluge/json" \
  -e PUFF_DELUGE_PASSWORD="your_password" \
  -e PUFF_RETENTION="P14D" \
  --name puff \
  ghcr.io/lukfisz/puff:latest
```

Run with qBittorrent:
```bash
docker run -d \
  -e PUFF_CRON_SCHEDULE="0 0 * * *" \
  -e PUFF_QBITTORRENT_URL="http://url_or_ip_to_qbittorrent:8080" \
  -e PUFF_QBITTORRENT_USERNAME="admin" \
  -e PUFF_QBITTORRENT_PASSWORD="your_password" \
  -e PUFF_RETENTION="P14D" \
  --name puff \
  ghcr.io/lukfisz/puff:latest
```

Run once and exit:
```bash
docker run --rm \
  -e PUFF_RUN_ONCE="true" \
  -e PUFF_DELUGE_URL="http://url_or_ip_to_deluge/json" \
  -e PUFF_DELUGE_PASSWORD="your_password" \
  -e PUFF_RETENTION="P14D" \
  --name puff-run-once \
  ghcr.io/lukfisz/puff:latest
```

### Docker Compose Example

Basic example with Deluge:

```yaml
services:
  puff:
    image: ghcr.io/lukfisz/puff:latest
    container_name: puff
    environment:
      PUFF_CRON_SCHEDULE: "0 0 * * *"
      PUFF_DELUGE_URL: "http://url_or_ip_to_deluge/json"
      PUFF_DELUGE_PASSWORD: "your_password"
      PUFF_RETENTION: P14D
    restart: unless-stopped
```

Basic example with qBittorrent:

```yaml
services:
  puff:
    image: ghcr.io/lukfisz/puff:latest
    container_name: puff
    environment:
      PUFF_CRON_SCHEDULE: "0 0 * * *"
      PUFF_QBITTORRENT_URL: "http://url_or_ip_to_qbittorrent:8080"
      PUFF_QBITTORRENT_USERNAME: "admin"
      PUFF_QBITTORRENT_PASSWORD: "your_password"
      PUFF_RETENTION: P14D
    restart: unless-stopped
```

Complete example with additional options:

```yaml
services:
  puff:
    image: ghcr.io/lukfisz/puff:latest
    container_name: puff
    environment:
      # Required: Cron schedule for retention job
      - PUFF_CRON_SCHEDULE=0 0 * * *
      # Deluge configuration (optional if qBittorrent is configured)
      - PUFF_DELUGE_URL=http://deluge.lan/json
      - PUFF_DELUGE_PASSWORD=${DELUGE_PASSWORD}
      # qBittorrent configuration (optional if Deluge is configured)
      - PUFF_QBITTORRENT_URL=http://qbittorrent.lan:8080
      - PUFF_QBITTORRENT_USERNAME=admin
      - PUFF_QBITTORRENT_PASSWORD=${QBIT_PASSWORD}
      # Optional: Retention period (default: P14D)
      - PUFF_RETENTION=P14D
      # Optional: Startup delay (default: 0s)
      - PUFF_START_DELAY=5s
      # Optional: HTTP client timeout (default: 2m0s)
      - PUFF_TORRENT_CLIENT_TIMEOUT=2m0s
      # Optional: Dry-run mode (default: false)
      - PUFF_DRY_RUN=false
      # Optional: Logging level (default: INFO)
      - PUFF_LOG_LEVEL=INFO
      # Optional: Disk free space threshold (e.g., 100GB, 1.5TB)
      - PUFF_DISK_FREE_SPACE_THRESHOLD=100GB
      # Optional: Path to disk to monitor (default: /mnt/puff/monitor)
      # - PUFF_DISK_PATH=/mnt/puff/monitor
      # Optional: Discord webhook url (default: "")
      - PUFF_DISCORD_WEBHOOK_URL="https://webhook.to.your.discord.channel"
      # Optional: Time zone (default: UTC)
      - TZ=Europe/Warsaw
    restart: unless-stopped
    # Uncomment to mount a volume for disk space monitoring
    # volumes:
    #   - /path/to/your/disk:/mnt/puff/monitor:ro
    # Optional: Health check
    healthcheck:
      test: [ "CMD", "pgrep", "-f", "puff" ]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

## How It Works

1. The application connects to configured torrent clients (Deluge via JSON-RPC, qBittorrent via Web API).
2. App flow:
    - If `PUFF_RUN_ONCE` is `true`:
      - The retention job runs immediately once for each configured client and the application exits.
    - If `PUFF_RUN_ONCE` is `false` (default):
      - Starts a job scheduler according to `PUFF_CRON_SCHEDULE`.
      - For each configured torrent client, in each cycle:
          - If `PUFF_DISK_FREE_SPACE_THRESHOLD` is set, checks if threshold was exceeded.
          - Retrieves a list of finished torrents.
          - Filters torrents that have exceeded the retention period.
          - Removes expired torrents along with their data (if `PUFF_DRY_RUN=false`).
          - Logs operation details, including freed disk space.
          - Sends a notification to Discord (if configured).

## Discord Notifications

Puff can send notifications to a Discord channel using a webhook. This is useful for monitoring the application's status and receiving timely alerts.

Notifications are sent for:
- **Job Summaries**: A summary of the retention job, including the number of removed torrents and the total space freed.
- **Errors**: Any errors or fatal events that occur during the application's execution.

To enable Discord notifications, you need to provide a webhook URL in the `PUFF_DISCORD_WEBHOOK_URL` environment variable. You can get a webhook URL from your Discord server's settings.

## Dry-Run Mode

In dry-run mode, the application only logs which torrents would be removed but does not perform the deletion operation.

## Development

### Project Structure

```
root/
├── Dockerfile              # Docker image
├── run.sh                  # Local execution script
├── main.go                 # Main application file
├── internal/
│   ├── config.go           # Configuration and environment variables
│   ├── context.go          # Application context with shared state
│   ├── job.go              # Job scheduler
│   ├── torrent.go          # TorrentClient and Torrent interfaces
│   ├── deluge.go           # Deluge JSON-RPC client
│   ├── qbittorrent.go      # qBittorrent Web API client
│   ├── discord.go          # Discord client
│   ├── retentionjob.go     # Torrent removal logic
│   ├── disk.go             # Disk free space checking
│   ├── utils.go            # Utility functions
│   └── preview.go          # Preview mode client
```

## Dependencies

- [gocron](https://github.com/go-co-op/gocron) - GoCron is a job scheduling package which lets you run Go functions periodically at pre-determined times using a simple, human-friendly syntax.
- [envconfig](https://github.com/kelseyhightower/envconfig) - A Go library for managing configuration data from environment variables.
- [log](https://github.com/charmbracelet/log) - A fancy logger for Go with stylish formatting and levels.
- [uuid](https://github.com/google/uuid) - A Go package for working with UUIDs (Universally Unique Identifiers).
- [date](https://github.com/rickb777/date) - A Go package for working with dates, times and ranges.
- [resty](https://resty.dev/) - Simple HTTP and REST client library for Go.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

lukFisz
