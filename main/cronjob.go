package main

import (
	"slices"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

func NewScheduler(config AppConfig) gocron.Scheduler {
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(config.Location()),
		gocron.WithStopTimeout(1*time.Minute),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	scheduler.Start()
	return scheduler
}

func ScheduleCronjob(ctx appContext, jobName string, job func()) gocron.Job {
	return scheduleJob(
		gocron.CronJob(ctx.AppConfig.Cron, true),
		*ctx.Scheduler,
		jobName,
		job,
		gocron.WithEventListeners(beforeJob(), afterJob(&ctx)),
	)
}

func NewOneTimeJob(jobName string, job func(), ctx appContext) gocron.Job {
	return scheduleJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()),
		*ctx.Scheduler,
		jobName,
		job,
		gocron.WithEventListeners(beforeJob(), afterJob(&ctx)),
		gocron.WithTags("run-once-job"),
	)
}

func scheduleJob(
	jobDef gocron.JobDefinition,
	scheduler gocron.Scheduler,
	jobName string,
	job func(),
	opts ...gocron.JobOption,
) gocron.Job {
	cronjob, err := scheduler.NewJob(
		jobDef,
		gocron.NewTask(job),
		append(
			opts,
			gocron.WithName(jobName),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		)...,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cronjob
}

func beforeJob() gocron.EventListener {
	return gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
		log.Info("running", "job", jobName)
	})
}

func afterJob(ctx *appContext) gocron.EventListener {
	return gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
		for _, j := range (*ctx.Scheduler).Jobs() {
			if j.ID() == jobID {
				if slices.Contains(j.Tags(), "run-once-job") {
					log.Info("finished", "job", jobName)
					go func() { *ctx.ShutdownChan <- Shutdown{} }()
					return
				}
				log.Info("finished", "job", jobName, "next run", GetNextRun(j))
				return
			}
		}
		log.Info("finished", "job", jobName)
	})
}

func GetNextRun(job gocron.Job) time.Time {
	nextRun, err := job.NextRun()
	if err != nil {
		log.Fatal(err.Error())
	}
	return nextRun
}
