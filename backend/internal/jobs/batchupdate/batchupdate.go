package batchupdate

import (
	"github.com/htchan/WebHistory/internal/jobs/websiteupdate"
	"github.com/htchan/WebHistory/internal/repository"
)

// TODO: add missing testcases
func Setup(rpo repository.Repostory, websiteUpdateScheduler *websiteupdate.Scheduler) *Scheduler {
	batchUpdateJob := NewJob(rpo, websiteUpdateScheduler)
	scheduler := NewScheduler(batchUpdateJob)

	return scheduler
}
