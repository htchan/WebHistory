package websiteupdate

import (
	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/repository"
)

// TODO: add missing testcases
func Setup(rpo repository.Repostory, conf *config.WorkerBinConfig) *Scheduler {
	websiteUpdateJob := NewJob(rpo, conf.WebsiteUpdateSleepInterval)
	scheduler := NewScheduler(websiteUpdateJob, conf)

	return scheduler
}
