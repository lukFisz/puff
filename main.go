package main

import (
	"encoding/json"
	"fmt"
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
(__)  \____/(__)  (__)  
by lukFisz
`

type appContext struct {
	Scheduler gocron.Scheduler
}

func main() {
	fmt.Print(appBanner)
	log.Print("started")

	config := GetConfig()

	setLogLevel(config)
	logConfig(config)
	delayAppStart(config)

	orchestrateExecution(
		func() appContext { return runApp(config) },
		gracefulShutdown,
	)
}

func setLogLevel(config AppConfig) {
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
	maskedConfig := config
	maskedConfig.DelugePassword = "***"
	prettyConfig, _ := json.MarshalIndent(maskedConfig, "", "  ")
	log.Info("init", "app config", string(prettyConfig))
}

func orchestrateExecution(executeLogic func() appContext, cleanUp func(appContext)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
	executionContext := executeLogic()
	<-sigs
	cleanUp(executionContext)
}

func runApp(config AppConfig) appContext {
	delugeClient := NewDelugeClient(config.DelugeUrl, config.DelugePassword)
	if err := delugeClient.Login(); err != nil {
		log.Error("deluge client", "err", err)
	} else {
		log.Info("deluge client", "authentication", "successful")
	}
	scheduler := NewScheduler(config)
	AddCronjob(
		scheduler,
		"Deluge torrent retention",
		config,
		func() { removeExpiredTorrents(*delugeClient, config) },
	)
	return appContext{Scheduler: scheduler}
}

func gracefulShutdown(executionContext appContext) {
	err := executionContext.Scheduler.Shutdown()
	if err != nil {
	}
	log.Info("shut down")
}
