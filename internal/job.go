package internal

import (
	"slices"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

const runOnceTag = "run-once-job"
const previewJobTag = "preview-job"

func NewScheduler(config AppConfig) gocron.Scheduler {
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(config.Location()),
		gocron.WithStopTimeout(time.Minute),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	scheduler.Start()
	return scheduler
}

func ScheduleCronjob(jobName string, job func(), ctx *AppContext) gocron.Job {
	options := make([]gocron.JobOption, 0)
	options = append(options, gocron.WithEventListeners(beforeJob(), afterJob(ctx)))
	if ctx.AppConfig.Preview {
		options = append(options, gocron.JobOption(gocron.WithStartImmediately()))
		options = append(options, gocron.WithTags(previewJobTag))
	}
	return scheduleJob(
		gocron.CronJob(ctx.AppConfig.Cron, true),
		*ctx.Scheduler,
		jobName,
		job,
		*ctx.AppConfig,
		options...,
	)
}

func NewOneTimeJob(jobName string, job func(), ctx *AppContext) gocron.Job {
	return scheduleJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartImmediately()),
		*ctx.Scheduler,
		jobName,
		job,
		*ctx.AppConfig,
		gocron.WithEventListeners(beforeJob(), afterJob(ctx)),
		gocron.WithTags(runOnceTag),
	)
}

func scheduleJob(
	jobDef gocron.JobDefinition,
	scheduler gocron.Scheduler,
	jobName string,
	job func(),
	appCfg AppConfig,
	opts ...gocron.JobOption,
) gocron.Job {
	if appCfg.Preview {
		time.Sleep(2 * time.Second)
	}

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

	if appCfg.Preview {
		log.Info("scheduled", "job", jobName, "next run", time.Now().In(appCfg.Location()))
	} else {
		log.Info("scheduled", "job", jobName, "next run", getNextRun(cronjob))
	}

	return cronjob
}

func beforeJob() gocron.EventListener {
	return gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
		log.Info("running", "job", jobName)
	})
}

func afterJob(ctx *AppContext) gocron.EventListener {
	return gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
		for _, j := range (*ctx.Scheduler).Jobs() {
			if j.ID() == jobID {
				if slices.Contains(j.Tags(), runOnceTag) {
					log.Info("finished", "job", jobName)
					*ctx.ShutdownChan <- true
					return
				}
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
