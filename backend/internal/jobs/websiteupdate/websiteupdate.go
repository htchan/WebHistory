package websiteupdate

import (
	"time"

	"github.com/htchan/WebHistory/internal/repository"
)

func Setup(rpo repository.Repostory, sleepInterval time.Duration) *Scheduler {
	websiteUpdateJob := NewJob(rpo, sleepInterval)
	scheduler := NewScheduler(websiteUpdateJob)

	return scheduler
}
