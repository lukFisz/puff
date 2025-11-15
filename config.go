package main

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Cron     string `envconfig:"BRUSH_CRON_SCHEDULE" required:"true"`
	TimeZone string `envconfig:"TZ" default:"Europe/Warsaw"`
}

func getConfig() config {
	var config config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return config
}

func (config config) location() *time.Location {
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		log.Fatal(err.Error())
	}
	return location
}
