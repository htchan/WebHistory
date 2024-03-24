package updatewebsite

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/htchan/WebHistory/internal/service"
	"github.com/htchan/goworkers"
	"github.com/htchan/goworkers/stream/redis"
	"github.com/redis/rueidis/rueidiscompat"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var (
	ErrInvalidParams = errors.New("invalid params")
)

type UpdateWebsiteTask struct {
	stream        *redis.RedisStream
	ch            chan interface{}
	cancel        func()
	lock          sync.Mutex
	rpo           repository.Repostory
	sleepInterval time.Duration
}

var _ goworkers.Task = (*UpdateWebsiteTask)(nil)

func NewUpdateWebsiteTask(stream *redis.RedisStream, rpo repository.Repostory, sleepInterval time.Duration) *UpdateWebsiteTask {
	return &UpdateWebsiteTask{stream: stream, rpo: rpo, sleepInterval: sleepInterval}
}

func (task *UpdateWebsiteTask) Subscribe(ctx context.Context, msgCh chan goworkers.Msg) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	task.cancel = cancel
	task.ch = make(chan interface{})

	// parse message to specific format and send it to worker pools
	go func() {
		for data := range task.ch {
			task.lock.Lock()
			redisMsg := data.(rueidiscompat.XMessage)
			msg := goworkers.Msg{
				TaskName: redisMsg.Values["task_name"].(string),
				Params:   data,
			}

			msgCh <- msg
		}
	}()

	// subscribe message from redis
	return task.stream.Subscribe(cancelCtx, task.ch)
}

func (task *UpdateWebsiteTask) Unsubscribe() error {
	// stop read from redis
	if task.cancel == nil {
		return nil
	}

	task.cancel()
	task.cancel = nil

	// stop pushing message from redis to worker pool
	close(task.ch)

	return nil
}

func (task *UpdateWebsiteTask) Publish(ctx context.Context, params interface{}) error {
	website, ok := params.(model.Website)
	if !ok {
		return ErrInvalidParams
	}

	encodedWebsite, err := json.Marshal(website)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"task_name": task.Name(),
		"web":       string(encodedWebsite),
	}

	return task.stream.Publish(ctx, data)
}

func (task *UpdateWebsiteTask) Execute(ctx context.Context, p interface{}) error {
	defer task.lock.Unlock()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	params, ok := p.(rueidiscompat.XMessage)
	if !ok {
		return ErrInvalidParams
	}

	tr := otel.Tracer("htchan/WebHistory/update-jobs")
	updateCtx, updateSpan := tr.Start(ctx, "Update Website")
	defer updateSpan.End()

	updateSpan.SetAttributes(attribute.String("task_name", params.Values["task_name"].(string)))

	var web model.Website
	jsonErr := json.Unmarshal([]byte(params.Values["web"].(string)), &web)
	if jsonErr != nil {
		return ErrInvalidParams
	}

	err := service.Update(updateCtx, task.rpo, &web)

	_, sleepSpan := tr.Start(updateCtx, "Sleep After Update")
	defer sleepSpan.End()
	time.Sleep(task.sleepInterval)

	if err != nil {
		updateSpan.RecordError(err)
		updateSpan.SetStatus(codes.Error, err.Error())
		return err
	}

	task.stream.Acknowledge(updateCtx, params.ID)

	return nil
}

func (task *UpdateWebsiteTask) Name() string {
	return "update_website_task"
}
