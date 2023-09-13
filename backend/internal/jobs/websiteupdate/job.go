package websiteupdate

import (
	"context"
	"time"

	"github.com/htchan/WebHistory/internal/executor"
	"github.com/htchan/WebHistory/internal/jobs"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/htchan/WebHistory/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// TODO: add missing testcases
type Job struct {
	rpo           repository.Repostory
	sleepInterval time.Duration
}

var _ executor.Job = (*Job)(nil)

func NewJob(rpo repository.Repostory, sleepInterval time.Duration) *Job {
	return &Job{
		rpo:           rpo,
		sleepInterval: sleepInterval,
	}
}

func (job *Job) Execute(ctx context.Context, p interface{}) error {
	params, ok := p.(Params)
	if !ok {
		return jobs.ErrInvalidParams
	}

	defer params.executionLock.Unlock()

	tr := otel.Tracer("htchan/WebHistory/update-jobs")
	ctx, span := tr.Start(ctx, "Update Website")
	defer span.End()
	span.SetAttributes(params.web.OtelAttributes()...)
	span.SetAttributes(attribute.String("job_uuid", ctx.Value("job_uuid").(string)))

	service.Update(ctx, job.rpo, params.web)
	time.Sleep(job.sleepInterval)

	return nil
}
