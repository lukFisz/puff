package internal

import (
	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
)

type AppContext struct {
	Scheduler     *gocron.Scheduler
	DiscordClient *DiscordClient
	TorrentClient *TorrentClient
	AppConfig     *AppConfig
	ShutdownChan  *chan bool
}

func NewAppContext(cfg AppConfig) *AppContext {
	if cfg.Cron == "" {
		log.Fatal("CRON SCHEDULE cannot be empty")
	}

	var torrentClient TorrentClient
	if cfg.Preview {
		torrentClient = NewPreviewTorrentClient()
	} else {
		torrentClient = NewDelugeClient(cfg)
	}
	discordClient := NewDiscordClient(cfg)
	scheduler := NewScheduler(cfg)

	shutdowns := make(chan bool)
	return &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		TorrentClient: &torrentClient,
		AppConfig:     &cfg,
		ShutdownChan:  &shutdowns,
	}
}
