package main

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/rickb777/date/period"
)

type AppConfig struct {
	Cron                string `envconfig:"PUFF_CRON_SCHEDULE" required:"true"`
	DelugeUrl           string `envconfig:"PUFF_DELUGE_URL" required:"true"`
	DelugePassword      string `envconfig:"PUFF_DELUGE_PASSWORD" required:"true"`
	DelugeClientTimeout string `envconfig:"PUFF_DELUGE_CLIENT_TIMEOUT" default:"2m0s"`
	Retention           string `envconfig:"PUFF_RETENTION" default:"P14D"`
	StartDelay          string `envconfig:"PUFF_START_DELAY" default:"0s"`
	DryRun              bool   `envconfig:"PUFF_DRY_RUN" default:"false"`
	LogLevel            string `envconfig:"PUFF_LOG_LEVEL" default:"INFO"`
	DiscordWebhookUrl   string `envconfig:"PUFF_DISCORD_WEBHOOK_URL" default:""`
	TimeZone            string `envconfig:"TZ" default:"UTC"`
	Version             string `envconfig:"PUFF_CURRENT_VERSION" default:""`
}

func GetConfig() AppConfig {
	var config AppConfig
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return config
}

func (config AppConfig) ParseValidation() {
	config.Location()
	config.RetentionInSeconds()
	config.Location()
	config.StartDelayDuration()
	config.DelugeClientTimeoutDuration()
}

func (config AppConfig) Location() *time.Location {
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		log.Fatal(err.Error())
	}
	return location
}

func (config AppConfig) RetentionInSeconds() int {
	period, err := period.Parse(config.Retention)
	if err != nil {
		log.Fatal(err.Error())
	}
	dur := period.DurationApprox()
	return int(dur.Seconds())
}

func (config AppConfig) StartDelayDuration() time.Duration {
	duration, err := time.ParseDuration(config.StartDelay)
	if err != nil {
		log.Fatal(err.Error())
	}
	return duration
}

func (config AppConfig) DelugeClientTimeoutDuration() time.Duration {
	duration, err := time.ParseDuration(config.DelugeClientTimeout)
	if err != nil {
		log.Fatal(err.Error())
	}
	return duration
}
