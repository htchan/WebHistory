package batchupdate

import (
	"time"

	"github.com/htchan/WebHistory/internal/executor"
	"github.com/htchan/WebHistory/internal/jobs"
)

// TODO: add missing testcases
// TODO: put this into websiteupdate
type Scheduler struct {
	job     *Job
	stop    chan struct{}
	jobChan executor.JobTrigger
}

var _ jobs.Scheduler = (*Scheduler)(nil)

func NewScheduler(job *Job) *Scheduler {
	return &Scheduler{
		job:     job,
		stop:    make(chan struct{}),
		jobChan: make(executor.JobTrigger, 1),
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
			lastRunTime = time.Now()

			// send job to executor
			scheduler.jobChan <- &executor.JobExec{
				Job:    scheduler.job,
				Params: Params{},
			}
		}
	}
}

func (scheduler *Scheduler) Stop() error {
	scheduler.stop <- struct{}{}
	close(scheduler.stop)
	close(scheduler.jobChan)

	return nil
}

func (scheduler *Scheduler) Publisher() executor.JobTrigger {
	return scheduler.jobChan
}

// run at specific time at every friday
func calculateNexRunTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)

	nDaysLater := int(time.Friday - now.Weekday())
	if nDaysLater <= 1 {
		nDaysLater += 7
	}

	result := now.AddDate(0, 0, nDaysLater).Add(4 * time.Hour)

	return result
}
