package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/muesli/termenv"
)

const appBanner string = ` ____  _  _  ____  ____ 
(  _ \/ )( \(  __)(  __)
 ) __/) \/ ( ) _)  ) _) 
(__)  \____/(__)  (__) %s
by lukFisz`

type Shutdown struct {
}

type appContext struct {
	Scheduler     *gocron.Scheduler
	DiscordClient *DiscordClient
	DelugeClient  *DelugeClient
	AppConfig     *AppConfig
	ShutdownChan  *chan Shutdown
}

func newAppContext(cfg AppConfig) appContext {
	if cfg.Cron == "" {
		log.Fatal("CRON SCHEDULE cannot be empty")
	}

	delugeClient := NewDelugeClient(cfg.DelugeUrl, cfg.DelugePassword, cfg.DelugeClientTimeoutDuration())
	delugeClient.CheckConnection()

	discordClient := NewDiscordClient(cfg)
	scheduler := NewScheduler(cfg)

	shutdowns := make(chan Shutdown)
	return appContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		DelugeClient:  delugeClient,
		AppConfig:     &cfg,
		ShutdownChan:  &shutdowns,
	}
}

func main() {
	config := GetConfig()
	config.ParseValidation()

	initLogger(config)

	initMessage(config)
	logConfig(config)

	ctx := newAppContext(config)

	delayAppStart(config)

	orchestrateExecution(
		func() appContext { return runApp(ctx) },
		gracefulShutdown,
	)
}

func initMessage(config AppConfig) {
	logger := log.New(os.Stdout)
	logger.SetReportTimestamp(false)
	logger.SetReportCaller(false)
	logger.Print(fmt.Sprintf(appBanner, config.Version))
	log.Print("started")
}

func initLogger(config AppConfig) {
	multi := io.MultiWriter(os.Stdout, NewDiscordClient(config))
	log.SetOutput(multi)
	log.SetColorProfile(termenv.TrueColor)
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal("log level", "err", err)
	}
	log.SetLevel(level)
}

func delayAppStart(config AppConfig) {
	startDelayDuration := config.StartDelayDuration()
	if startDelayDuration.Seconds() > 0 {
		log.Info("start delay", "duration", startDelayDuration)
		time.Sleep(startDelayDuration)
	}
}

func logConfig(config AppConfig) {
	strConfig := fmt.Sprintf(`  CRON_SCHEDULE: %s
  DELUGE_URL: %s
  DELUGE_PASSWORD: ****
  DELUGE_CLIENT_TIMEOUT: %s
  RETENTION: %s
  START_DELAY: %s
  DRY_RUN: %t
  LOG_LEVEL: %s
  DISCORD_WEBHOOK_URL: ****
  RUN_ONCE: %t
  TZ: %s`,
		config.Cron,
		config.DelugeUrl,
		config.DelugeClientTimeout,
		config.Retention,
		config.StartDelay,
		config.DryRun,
		config.LogLevel,
		config.RunOnce,
		config.TimeZone,
	)
	log.Info("init", "app config", strConfig)
}

func orchestrateExecution(executeLogic func() appContext, cleanUp func(appContext)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
	ctx := executeLogic()
	select {
	case <-sigs:
	case <-*ctx.ShutdownChan:
	}
	cleanUp(ctx)
}

func runApp(ctx appContext) appContext {
	jobName := "Deluge torrent retention"
	var job gocron.Job
	delTorrentsJob := func() { RemoveExpiredTorrents(*ctx.DelugeClient, ctx.DiscordClient, *ctx.AppConfig) }
	if ctx.AppConfig.RunOnce {
		job = NewOneTimeJob(jobName, func() {
			delTorrentsJob()
			*ctx.ShutdownChan <- Shutdown{}
		}, ctx)
	} else {
		job = ScheduleCronjob(ctx, jobName, delTorrentsJob)
		log.Info("scheduled", "job", jobName, "next run", GetNextRun(job))
	}

	return ctx
}

func gracefulShutdown(ctx appContext) {
	err := (*ctx.Scheduler).Shutdown()
	if err != nil {
		log.Fatal("scheduler", "err", err)
	}
	log.Print("shut down")
}
