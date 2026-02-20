package internal

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/kelseyhightower/envconfig"
	"github.com/rickb777/date/period"
	"github.com/rs/zerolog/log"
)

type AppConfig struct {
	Cron                   string  `envconfig:"PUFF_CRON_SCHEDULE"`
	QbitTorrentUrl         string  `envconfig:"PUFF_QBITTORRENT_URL" required:"false"`
	QbitTorrentPassword    string  `envconfig:"PUFF_QBITTORRENT_PASSWORD" required:"false"`
	QbitTorrentUsername    string  `envconfig:"PUFF_QBITTORRENT_USERNAME" required:"false"`
	DelugeUrl              string  `envconfig:"PUFF_DELUGE_URL" required:"false"`
	DelugePassword         string  `envconfig:"PUFF_DELUGE_PASSWORD" required:"false"`
	TorrentClientTimeout   string  `envconfig:"PUFF_TORRENT_CLIENT_TIMEOUT" default:"2m0s"`
	Retention              string  `envconfig:"PUFF_RETENTION" default:"P14D"`
	StartDelay             string  `envconfig:"PUFF_START_DELAY" default:"0s"`
	DiskFreeSpaceThreshold *string `envconfig:"PUFF_DISK_FREE_SPACE_THRESHOLD"`
	DiskPath               string  `envconfig:"PUFF_DISK_PATH" default:"/mnt/puff/monitor"`
	DryRun                 bool    `envconfig:"PUFF_DRY_RUN" default:"false"`
	LogLevel               string  `envconfig:"PUFF_LOG_LEVEL" default:"INFO"`
	DiscordWebhookUrl      string  `envconfig:"PUFF_DISCORD_WEBHOOK_URL" default:""`
	RunOnce                bool    `envconfig:"PUFF_RUN_ONCE" default:"false"`
	TimeZone               string  `envconfig:"TZ" default:"UTC"`
	Version                string  `envconfig:"PUFF_CURRENT_VERSION" default:""`
	Preview                bool    `envconfig:"PUFF_PREVIEW_MODE" default:"false"`
}

func GetConfig() AppConfig {
	var config AppConfig
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal().Err(err).Msg("config error")
	}
	return config
}

func (config AppConfig) ParseValidation() {
	config.Location()
	config.RetentionInSeconds()
	config.StartDelayDuration()
	config.TorrentClientTimeoutDuration()
	config.DiskFreeSpaceThresholdInBytes()
	if config.DiskFreeSpaceThreshold != nil {
		_, err := config.FreeSpaceOfDiskPathInBytes()
		if err != nil {
			log.Fatal().Err(err).Msg("DISK_PATH validation")
		}
	}
}

func (config AppConfig) Location() *time.Location {
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		log.Fatal().Err(err).Msg("timezone error")
	}
	return location
}

func (config AppConfig) RetentionInSeconds() int {
	period, err := period.Parse(config.Retention)
	if err != nil {
		log.Fatal().Err(err).Msg("retention parse error")
	}
	dur := period.DurationApprox()
	return int(dur.Seconds())
}

func (config AppConfig) StartDelayDuration() time.Duration {
	duration, err := time.ParseDuration(config.StartDelay)
	if err != nil {
		log.Fatal().Err(err).Msg("start delay parse error")
	}
	return duration
}

func (config AppConfig) TorrentClientTimeoutDuration() time.Duration {
	duration, err := time.ParseDuration(config.TorrentClientTimeout)
	if err != nil {
		log.Fatal().Err(err).Msg("torrent client timeout parse error")
	}
	return duration
}

func (config AppConfig) DiskFreeSpaceThresholdInBytes() *uint64 {
	if config.DiskFreeSpaceThreshold == nil {
		return nil
	}
	size, err := humanize.ParseBytes(*config.DiskFreeSpaceThreshold)
	if err != nil {
		log.Fatal().Err(err).Msg("disk free space threshold parse error")
	}
	return &size
}

func (config AppConfig) FreeSpaceOfDiskPathInBytes() (*uint64, error) {
	freeSpace, err := DiskFreeSpaceInBytes(config.DiskPath)
	if err != nil {
		return nil, err
	}
	return freeSpace, nil
}
