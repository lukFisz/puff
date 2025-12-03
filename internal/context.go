package internal

import (
	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
)

type AppContext struct {
	Scheduler     *gocron.Scheduler
	DiscordClient *DiscordClient
	DelugeClient  *DelugeClient
	AppConfig     *AppConfig
	ShutdownChan  *chan bool
}

func NewAppContext(cfg AppConfig) *AppContext {
	if cfg.Cron == "" {
		log.Fatal("CRON SCHEDULE cannot be empty")
	}

	delugeClient := NewDelugeClient(cfg)
	delugeClient.CheckConnection()

	discordClient := NewDiscordClient(cfg)
	scheduler := NewScheduler(cfg)

	shutdowns := make(chan bool)
	return &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		DelugeClient:  delugeClient,
		AppConfig:     &cfg,
		ShutdownChan:  &shutdowns,
	}
}
