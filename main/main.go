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
)

const appBanner string = ` ____  _  _  ____  ____ 
(  _ \/ )( \(  __)(  __)
 ) __/) \/ ( ) _)  ) _) 
(__)  \____/(__)  (__) %s
by lukFisz`

type appContext struct {
	Scheduler gocron.Scheduler
}

func main() {
	config := GetConfig()
	initLogger(config)

	initMessage()
	config.ParseValidation()
	logConfig(config)
	delayAppStart(config)

	orchestrateExecution(
		func() appContext { return runApp(config) },
		gracefulShutdown,
	)
}

func initMessage() {
	version := os.Getenv("PUFF_CURRENT_VERSION")
	logger := log.New(os.Stdout)
	logger.SetReportTimestamp(false)
	logger.SetReportCaller(false)
	logger.Print(fmt.Sprintf(appBanner, version))
	log.Print("started")
}

func initLogger(config AppConfig) {
	multi := io.MultiWriter(os.Stdout, NewDiscordClient(config))
	log.SetOutput(multi)
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
  TZ: %s`,
		config.Cron,
		config.DelugeUrl,
		config.DelugeClientTimeout,
		config.Retention,
		config.StartDelay,
		config.DryRun,
		config.LogLevel,
		config.TimeZone,
	)
	log.Info("init", "app config", strConfig)
}

func orchestrateExecution(executeLogic func() appContext, cleanUp func(appContext)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
	executionContext := executeLogic()
	<-sigs
	cleanUp(executionContext)
}

func runApp(config AppConfig) appContext {
	delugeClient := NewDelugeClient(config.DelugeUrl, config.DelugePassword, config.DelugeClientTimeoutDuration())
	if err := delugeClient.Login(); err != nil {
		log.Error("deluge client", "err", err)
	} else {
		log.Info("deluge client", "authentication", "successful")
	}
	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	AddCronjob(
		scheduler,
		"Deluge torrent retention",
		config,
		func() { removeExpiredTorrents(*delugeClient, discordClient, config) },
	)
	return appContext{Scheduler: scheduler}
}

func gracefulShutdown(executionContext appContext) {
	err := executionContext.Scheduler.Shutdown()
	if err != nil {
	}
	log.Print("shut down")
}
