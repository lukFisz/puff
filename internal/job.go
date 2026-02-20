package internal

import (
	"slices"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const runOnceTag = "run-once-job"
const previewJobTag = "preview-job"

func NewScheduler(config AppConfig) gocron.Scheduler {
	scheduler, err := gocron.NewScheduler(
		gocron.WithLocation(config.Location()),
		gocron.WithStopTimeout(time.Minute),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("scheduler init")
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
		log.Fatal().Err(err).Msg("schedule job")
	}

	if appCfg.Preview {
		log.Info().Str("job", jobName).Time("next run", time.Now().In(appCfg.Location())).Msg("scheduled")
	} else {
		log.Info().Str("job", jobName).Time("next run", getNextRun(cronjob)).Msg("scheduled")
	}

	return cronjob
}

func beforeJob() gocron.EventListener {
	return gocron.BeforeJobRuns(func(jobID uuid.UUID, jobName string) {
		log.Info().Str("job", jobName).Msg("running")
	})
}

func afterJob(ctx *AppContext) gocron.EventListener {
	return gocron.AfterJobRuns(func(jobID uuid.UUID, jobName string) {
		for _, j := range (*ctx.Scheduler).Jobs() {
			if j.ID() == jobID {
				if slices.Contains(j.Tags(), runOnceTag) {
					log.Info().Str("job", jobName).Msg("finished")
					*ctx.ShutdownChan <- true
					return
				}
				log.Info().Str("job", jobName).Time("next run", getNextRun(j)).Msg("finished")
				return
			}
		}
		log.Info().Str("job", jobName).Msg("finished")
	})
}

func getNextRun(job gocron.Job) time.Time {
	nextRun, err := job.NextRun()
	if err != nil {
		log.Fatal().Err(err).Msg("get next run")
	}
	return nextRun
}
