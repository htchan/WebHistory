package batchupdate

import (
	"context"
	"fmt"

	"github.com/htchan/WebHistory/internal/executor"
	"github.com/htchan/WebHistory/internal/jobs"
	"github.com/htchan/WebHistory/internal/jobs/websiteupdate"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
)

type Job struct {
	rpo repository.Repostory

	websiteUpdateScheduler *websiteupdate.Scheduler
}

var _ executor.Job = (*Job)(nil)

func NewJob(rpo repository.Repostory, scheduler *websiteupdate.Scheduler) *Job {
	return &Job{
		rpo:                    rpo,
		websiteUpdateScheduler: scheduler,
	}
}

func (job *Job) Execute(ctx context.Context, p interface{}) error {
	_, ok := p.(Params)
	if !ok {
		return jobs.ErrInvalidParams
	}

	tr := otel.Tracer("htchan/WebHistory/batch-update")
	ctx, span := tr.Start(ctx, "Execute")
	defer span.End()

	// list all websites
	webs, err := job.rpo.FindWebsites()
	if err != nil {
		zerolog.Ctx(ctx).
			Error().
			Err(err).
			Msg("failed to list websites")
		return fmt.Errorf("failed to list websites: %w", err)
	}

	for _, web := range webs {
		// deploy job to update website
		err := job.websiteUpdateScheduler.DeployJob(&web)
		if err != nil {
			zerolog.Ctx(ctx).
				Error().
				Str("website", web.URL).
				Err(err).
				Msg("failed to deploy job to update website")
		}
	}

	return nil
}
