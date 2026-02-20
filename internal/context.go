package internal

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

type TorrentClientJob struct {
	TorrentType   string
	TorrentClient TorrentClient
}

type AppContext struct {
	Scheduler     *gocron.Scheduler
	DiscordClient *DiscordClient
	Jobs          *[]TorrentClientJob
	AppConfig     *AppConfig
	ShutdownChan  *chan bool
}

func NewAppContext(cfg AppConfig) *AppContext {
	if cfg.Cron == "" {
		log.Fatal().Msg("CRON SCHEDULE cannot be empty")
	}

	job := make([]TorrentClientJob, 0)
	if cfg.Preview {
		job = append(job, TorrentClientJob{TorrentType: "preview", TorrentClient: NewPreviewTorrentClient()})
	} else {
		if cfg.DelugeUrl != "" {
			job = append(job, TorrentClientJob{TorrentType: "deluge", TorrentClient: NewDelugeClient(cfg)})
			log.Info().Msg("Deluge client enabled")
		}
		if cfg.QbitTorrentUrl != "" {
			job = append(job, TorrentClientJob{TorrentType: "qbittorrent", TorrentClient: NewQbitTorrentClient(cfg)})
			log.Info().Msg("Qbittorrent client enabled")
		}
	}

	discordClient := NewDiscordClient(cfg)

	scheduler := NewScheduler(cfg)

	shutdowns := make(chan bool)
	return &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		Jobs:          &job,
		AppConfig:     &cfg,
		ShutdownChan:  &shutdowns,
	}
}
