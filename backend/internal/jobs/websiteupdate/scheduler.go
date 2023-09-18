package websiteupdate

import (
	"sync"
	"time"

	"github.com/htchan/WebHistory/internal/executor"
	"github.com/htchan/WebHistory/internal/jobs"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/rs/zerolog/log"
)

// TODO: add missing testcases
type Scheduler struct {
	job         *Job
	stop        chan struct{}
	jobChan     executor.JobTrigger
	hostLocks   map[string]*sync.Mutex
	publisherWg sync.WaitGroup
}

func NewScheduler(job *Job) *Scheduler {
	return &Scheduler{
		job:       job,
		stop:      make(chan struct{}),
		jobChan:   make(executor.JobTrigger, 1),
		hostLocks: make(map[string]*sync.Mutex),
	}
}

func (scheduler *Scheduler) Start() {
	// calculate next run time
	lastRunTime := time.Time{}
	for {
		select {
		case <-scheduler.stop:
			return
		case <-time.NewTimer(time.Until(calculateNexRunTime(lastRunTime))).C:
			lastRunTime = time.Now().UTC()
			scheduler.batchDeployUpdateJob()
		}
	}
}

// run at specific time at every friday
func calculateNexRunTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)

	nDaysLater := int(time.Friday - now.Weekday())
	if nDaysLater <= 1 {
		nDaysLater += 7
	}

	result := now.AddDate(0, 0, nDaysLater).Add(4 * time.Hour)

	return result
}

func (scheduler *Scheduler) batchDeployUpdateJob() {
	logger := log.With().
		Str("scheduler", "websiteupdate").
		Str("operation", "batch-update").
		Logger()

	logger.Info().Msg("batch update start")

	// list all websites
	webs, err := scheduler.job.rpo.FindWebsites()
	if err != nil {
		logger.Error().Err(err).Msg("failed to list websites")

		return
	}
	webs = []model.Website{
		{Title: "test", URL: "https://www.google.com"},
		{Title: "test2", URL: "https://www.google.com"},
	}

	for _, web := range webs {
		// deploy job to update website
		web := web
		err := scheduler.DeployJob(&web)
		if err != nil {
			logger.Error().Err(err).Str("website", web.URL).
				Msg("failed to deploy job to update website")
		}
	}
}

func (scheduler *Scheduler) Stop() error {
	close(scheduler.stop)
	scheduler.publisherWg.Wait()
	close(scheduler.jobChan)
	return nil
}

func (scheduler *Scheduler) Publisher() executor.JobTrigger {
	return scheduler.jobChan
}

func (scheduler *Scheduler) DeployJob(web *model.Website) error {
	// return error if the scheduler was stopped
	select {
	case <-scheduler.stop:
		return jobs.ErrSchedulerStopped
	default:
	}

	host := web.Host()

	// init executionLock
	lock, ok := scheduler.hostLocks[host]
	if !ok {
		lock = &sync.Mutex{}
		scheduler.hostLocks[host] = lock
	}

	scheduler.publisherWg.Add(1)
	go func() {
		defer scheduler.publisherWg.Done()

		lock.Lock()
		scheduler.jobChan <- &executor.JobExec{
			Job: scheduler.job,
			Params: Params{
				Web:           web,
				ExecutionLock: lock,
			},
		}
	}()

	return nil
}
