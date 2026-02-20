package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	puff "puff/internal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const appBanner string = ` ____  _  _  ____  ____
(  _ \/ )( \(  __)(  __)
 ) __/) \/ ( ) _)  ) _)
(__)  \____/(__)  (__) %s
by lukFisz`

func main() {
	config := puff.GetConfig()
	config.ParseValidation()

	initLogger(config)

	initMessage(config)
	logConfig(config)

	ctx := puff.NewAppContext(config)

	delayAppStart(config)

	orchestrateExecution(
		func() *puff.AppContext { return runApp(ctx) },
		gracefulShutdown,
	)
}

func initMessage(config puff.AppConfig) {
	fmt.Println(fmt.Sprintf(appBanner, config.Version))
	log.Info().Msg("started")
}

func initLogger(config puff.AppConfig) {
	zerolog.TimeFieldFormat = time.RFC3339
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	multi := io.MultiWriter(consoleWriter, puff.NewDiscordClient(config))
	level, err := zerolog.ParseLevel(strings.ToLower(config.LogLevel))
	if err != nil {
		log.Fatal().Err(err).Msg("log level")
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
}

func delayAppStart(config puff.AppConfig) {
	startDelayDuration := config.StartDelayDuration()
	if startDelayDuration.Seconds() > 0 {
		log.Info().Dur("duration", startDelayDuration).Msg("start delay")
		time.Sleep(startDelayDuration)
	}
}

func logConfig(config puff.AppConfig) {
	event := log.Info().
		Str("CRON_SCHEDULE", config.Cron).
		Str("TORRENT_CLIENT_TIMEOUT", config.TorrentClientTimeout).
		Str("RETENTION", config.Retention).
		Str("START_DELAY", config.StartDelay).
		Bool("DRY_RUN", config.DryRun).
		Str("LOG_LEVEL", config.LogLevel).
		Str("DISCORD_WEBHOOK_URL", "****").
		Bool("RUN_ONCE", config.RunOnce).
		Str("TZ", config.TimeZone)

	if config.DelugeUrl != "" {
		event = event.Str("DELUGE_URL", config.DelugeUrl)
	}

	if config.QbitTorrentUrl != "" {
		event = event.
			Str("QBITTORRENT_URL", config.QbitTorrentUrl).
			Str("QBITTORRENT_USERNAME", config.QbitTorrentUsername)
	}

	if config.DiskFreeSpaceThreshold != nil {
		event = event.
			Str("DISK_FREE_SPACE_THRESHOLD", *config.DiskFreeSpaceThreshold).
			Str("DISK_PATH", config.DiskPath)
	}

	event.Msg("init config")
}

func orchestrateExecution(executeLogic func() *puff.AppContext, cleanUp func(*puff.AppContext)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
	ctx := executeLogic()
	select {
	case <-sigs:
	case <-*ctx.ShutdownChan:
	}
	cleanUp(ctx)
}

func runApp(ctx *puff.AppContext) *puff.AppContext {
	for _, j := range *ctx.Jobs {
		j.TorrentClient.CheckConnection()
		jobName := j.TorrentType + " retention"
		delTorrentsJob := func() { puff.RemoveExpiredTorrents(ctx, j.TorrentClient) }
		if ctx.AppConfig.RunOnce {
			puff.NewOneTimeJob(jobName, delTorrentsJob, ctx)
		} else {
			puff.ScheduleCronjob(jobName, delTorrentsJob, ctx)
		}
	}
	return ctx
}

func gracefulShutdown(ctx *puff.AppContext) {
	err := (*ctx.Scheduler).Shutdown()
	if err != nil {
		log.Fatal().Err(err).Msg("scheduler")
	}
	log.Info().Msg("shut down")
}
