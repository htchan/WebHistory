package batchupdate

import (
	"github.com/htchan/WebHistory/internal/jobs/websiteupdate"
	"github.com/htchan/WebHistory/internal/repository"
)

func Setup(rpo repository.Repostory, websiteUpdateScheduler *websiteupdate.Scheduler) *Scheduler {
	batchUpdateJob := NewJob(rpo, websiteUpdateScheduler)
	scheduler := NewScheduler(batchUpdateJob)

	return scheduler
}
