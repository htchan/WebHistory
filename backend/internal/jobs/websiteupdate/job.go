package websiteupdate

import (
	"context"
	"runtime"
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

	defer params.ExecutionLock.Unlock()

	tr := otel.Tracer("htchan/WebHistory/update-jobs")

	updateCtx, updateSpan := tr.Start(ctx, "Update Website")
	defer updateSpan.End()

	updateSpan.SetAttributes(params.Web.OtelAttributes()...)
	updateSpan.SetAttributes(attribute.String("job_uuid", updateCtx.Value("job_uuid").(string)))

	service.Update(updateCtx, job.rpo, params.Web)

	_, sleepSpan := tr.Start(updateCtx, "Sleep After Update")
	defer sleepSpan.End()
	time.Sleep(job.sleepInterval)

	runtime.GC()

	return nil
}
