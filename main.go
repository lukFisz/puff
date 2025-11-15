package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	config := getConfig()

	log.Info("init", "app config", config)

	orchestrateExecution(
		func() appContext { return runApp(config) },
		gracefulShutdown,
	)
}

func orchestrateExecution(executeLogic func() appContext, cleanUp func(appContext)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
	executionContext := executeLogic()
	<-sigs
	cleanUp(executionContext)
}

func runApp(config config) appContext {
	scheduler := newScheduler(config)
	addCronjob(
		scheduler,
		"torrent retention",
		config,
		func() { log.Print("doing stuff...") },
	)
	return appContext{Scheduler: scheduler}
}

func gracefulShutdown(executionContext appContext) {
	err := executionContext.Scheduler.Shutdown()
	if err != nil {
	}
	log.Info("shut down")
}
