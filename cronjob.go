package main

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

func newScheduler(config config) gocron.Scheduler {
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(config.location()),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	scheduler.Start()
	return scheduler
}

func addCronjob(scheduler gocron.Scheduler, jobName string, config config, job func()) {
	cronjob, err := scheduler.NewJob(
		gocron.CronJob(config.Cron, true),
		gocron.NewTask(job),
		gocron.WithEventListeners(beforeJob(), afterJob(scheduler)),
		gocron.WithName(jobName),
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("scheduled", "job", jobName, "next run", getNextRun(cronjob))
}

func beforeJob() gocron.EventListener {
	return gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
		log.Info("running", "job", jobName)
	})
}

func afterJob(scheduler gocron.Scheduler) gocron.EventListener {
	return gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
		for _, j := range scheduler.Jobs() {
			if j.ID() == jobID {
				log.Info("finished", "job", jobName, "next run", getNextRun(j))
				return
			}
		}
		log.Info("finished", "job", jobName)
	})
}

func getNextRun(job gocron.Job) time.Time {
	nextRun, err := job.NextRun()
	if err != nil {
		log.Fatal(err.Error())
	}
	return nextRun
}
