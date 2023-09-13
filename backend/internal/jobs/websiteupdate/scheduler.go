package websiteupdate

import (
	"sync"

	"github.com/htchan/WebHistory/internal/executor"
	"github.com/htchan/WebHistory/internal/jobs"
	"github.com/htchan/WebHistory/internal/model"
)

type Scheduler struct {
	job         *Job
	stop        chan struct{}
	jobChan     executor.JobTrigger
	hostLocks   map[string]*sync.Mutex
	publisherWg *sync.WaitGroup
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

}

func (scheduler *Scheduler) Stop() error {
	//TODO: implement this
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
		hostLock := &sync.Mutex{}
		scheduler.hostLocks[host] = hostLock
	}

	scheduler.publisherWg.Add(1)
	go func() {
		defer scheduler.publisherWg.Done()

		lock.Lock()
		scheduler.jobChan <- &executor.JobExec{
			Job: scheduler.job,
			Params: Params{
				web:           web,
				executionLock: lock,
			},
		}
	}()

	return nil
}
