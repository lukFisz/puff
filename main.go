package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	puff "puff/internal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
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
	logger := log.New(os.Stdout)
	logger.SetReportTimestamp(false)
	logger.SetReportCaller(false)
	logger.Print(fmt.Sprintf(appBanner, config.Version))
	log.Print("started")
}

func initLogger(config puff.AppConfig) {
	multi := io.MultiWriter(os.Stdout, puff.NewDiscordClient(config))
	log.SetOutput(multi)
	log.SetColorProfile(termenv.TrueColor)
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal("log level", "err", err)
	}
	log.SetLevel(level)
}

func delayAppStart(config puff.AppConfig) {
	startDelayDuration := config.StartDelayDuration()
	if startDelayDuration.Seconds() > 0 {
		log.Info("start delay", "duration", startDelayDuration)
		time.Sleep(startDelayDuration)
	}
}

func logConfig(config puff.AppConfig) {
	var diskThreshold string
	if config.DiskFreeSpaceThreshold == nil {
		diskThreshold = "DISK_FREE_SPACE_TRESHOLD: NONE"
	} else {
		diskThreshold = fmt.Sprintf("DISK_FREE_SPACE_TRESHOLD: %s\n  DISK_PATH: %s", *config.DiskFreeSpaceThreshold, config.DiskPath)
	}
	strConfig := fmt.Sprintf(`  CRON_SCHEDULE: %s
  DELUGE_URL: %s
  QBITTORRENT_URL: %s
  QBITTORRENT_USERNAME: %s
  DELUGE_CLIENT_TIMEOUT: %s
  RETENTION: %s
  START_DELAY: %s
  DRY_RUN: %t
  LOG_LEVEL: %s
  DISCORD_WEBHOOK_URL: ****
  RUN_ONCE: %t
  %s
  TZ: %s`,
		config.Cron,
		config.DelugeUrl,
		config.QbitTorrentUrl,
		config.QbitTorrentUsername,
		config.TorrentClientTimeout,
		config.Retention,
		config.StartDelay,
		config.DryRun,
		config.LogLevel,
		config.RunOnce,
		diskThreshold,
		config.TimeZone,
	)
	log.Info("init", "app config", strConfig)
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
		log.Fatal("scheduler", "err", err)
	}
	log.Print("shut down")
}
