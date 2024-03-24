package updatewebsite

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/htchan/WebHistory/internal/repository/mockrepo"
	"github.com/htchan/WebHistory/internal/service"
	"github.com/htchan/goworkers"
	"github.com/htchan/goworkers/stream/redis"
	"github.com/redis/rueidis/rueidiscompat"
	"github.com/stretchr/testify/assert"
)

func Test_NewUpdateWebsiteTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stream        *redis.RedisStream
		rpo           repository.Repostory
		sleepInterval time.Duration
		want          *UpdateWebsiteTask
	}{
		{
			name:          "happy flow",
			stream:        &redis.RedisStream{},
			rpo:           nil,
			sleepInterval: time.Second,
			want: &UpdateWebsiteTask{
				stream:        &redis.RedisStream{},
				rpo:           nil,
				sleepInterval: time.Second,
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := NewUpdateWebsiteTask(test.stream, test.rpo, test.sleepInterval)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestUpdateWebsiteTask_Subscribe_Unsubscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                                string
		streamName, groupName, consumerName string
		getCtx                              func() context.Context
		publishWebsite                      model.Website
		ch                                  chan goworkers.Msg
		wantError                           error
		verifyMsg                           func(t *testing.T, msg chan goworkers.Msg)
	}{
		{
			name:         "happy flow",
			streamName:   "test/update-website-task/subscribe/happy-flow",
			groupName:    "test",
			consumerName: "test",
			getCtx: func() context.Context {
				return context.Background()
			},
			publishWebsite: model.Website{
				UUID:       "uuid",
				URL:        "url",
				Title:      "title",
				RawContent: "content",
				UpdateTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				Conf:       &config.WebsiteConfig{MaxDateLength: 1, Separator: "\n"},
			},
			ch:        make(chan goworkers.Msg),
			wantError: context.Canceled,
			verifyMsg: func(t *testing.T, msg chan goworkers.Msg) {
				for msg := range msg {
					assert.Equal(t, "update_website_task", msg.TaskName)
					assert.Equal(t, map[string]interface{}{
						"task_name": "update_website_task",
						"web":       `{"uuid":"uuid","url":"url","title":"title","raw_content":"content","update_time":"2022-01-01T00:00:00Z","Conf":{"Separator":"\n","MaxDateLength":1}}`,
					}, msg.Params.(rueidiscompat.XMessage).Values)
				}
			},
		},
		{
			name:         "cancelled context",
			streamName:   "test/update-website-task/subscribe/cancel",
			groupName:    "test",
			consumerName: "test",
			getCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			},
			publishWebsite: model.Website{},
			ch:             make(chan goworkers.Msg),
			wantError:      context.Canceled,
			verifyMsg:      func(t *testing.T, msg chan goworkers.Msg) {},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := test.getCtx()
			task := &UpdateWebsiteTask{
				stream: redis.NewRedisStream(
					cli, test.streamName, test.groupName, test.consumerName, time.Second,
				),
			}

			rueidiscompat.NewAdapter(cli).XGroupCreateMkStream(
				ctx, test.streamName, test.groupName, "0")

			task.Publish(ctx, test.publishWebsite)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := task.Subscribe(ctx, test.ch)
				assert.ErrorIs(t, err, test.wantError)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(50 * time.Millisecond)
				task.Unsubscribe()
				close(test.ch)
			}()

			test.verifyMsg(t, test.ch)
			wg.Wait()
		})
	}
}

func TestUpdateWebsiteTask_Publish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                                string
		streamName, groupName, consumerName string
		getCtx                              func() context.Context
		params                              model.Website
		wantResult                          map[string]interface{}
		wantError                           error
	}{
		{
			name:         "happy flow",
			streamName:   "test/update-website-task/publish/happy-flow",
			groupName:    "test",
			consumerName: "test",
			getCtx: func() context.Context {
				return context.Background()
			},
			params: model.Website{
				UUID:       "uuid",
				URL:        "url",
				Title:      "title",
				RawContent: "content",
				UpdateTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				Conf: &config.WebsiteConfig{
					MaxDateLength: 1,
					Separator:     "\n",
				},
			},
			wantResult: map[string]interface{}{
				"task_name": "update_website_task",
				"web":       `{"uuid":"uuid","url":"url","title":"title","raw_content":"content","update_time":"2022-01-01T00:00:00Z","Conf":{"Separator":"\n","MaxDateLength":1}}`,
			},
		},
		{
			name:         "cancelled context",
			streamName:   "test/update-website-task/publish/cancel",
			groupName:    "test",
			consumerName: "test",
			getCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			},
			params: model.Website{
				UUID:       "uuid",
				URL:        "url",
				Title:      "title",
				RawContent: "content",
				UpdateTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				Conf: &config.WebsiteConfig{
					MaxDateLength: 1,
					Separator:     "\n",
				},
			},
			wantError: context.Canceled,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := test.getCtx()
			task := &UpdateWebsiteTask{
				stream: redis.NewRedisStream(
					cli, test.streamName, test.groupName, test.consumerName, time.Second,
				),
			}

			err := task.Publish(test.getCtx(), test.params)
			assert.ErrorIs(t, err, test.wantError)

			if err != nil {
				return
			}

			stream, err := rueidiscompat.NewAdapter(cli).XRead(ctx, rueidiscompat.XReadArgs{
				Streams: []string{test.streamName, "0"},
				Count:   1,
				Block:   time.Millisecond,
			}).Result()
			assert.NoError(t, err)
			if assert.Equal(t, 1, len(stream)) {
				assert.Equal(t, test.wantResult, stream[0].Messages[0].Values)
			}
		})
	}
}

type MockClient struct {
	get func(string) (*http.Response, error)
}

func (m MockClient) Get(url string) (*http.Response, error) {
	return m.get(url)
}

func TestUpdateWebsiteTask_Execute(t *testing.T) {
	t.Parallel()

	mockRespWithDates := `<html><head>
	<title>new title</title>
	<dates><date>date-1</date><date>date-2</date>
	<date>date-3</date><date>date-4</date></dates>
	</head></html>`

	tests := []struct {
		name                                string
		streamName, groupName, consumerName string
		getCtx                              func() context.Context
		getRpo                              func(ctrl *gomock.Controller) repository.Repostory
		cli                                 service.HTTPClient
		params                              rueidiscompat.XMessage
		wantError                           error
	}{
		{
			getCtx: func() context.Context {
				return context.Background()
			},
			getRpo: func(ctrl *gomock.Controller) repository.Repostory {
				rpo := mockrepo.NewMockRepostory(ctrl)
				rpo.EXPECT().FindWebsiteSetting("google.com").
					Return(&model.WebsiteSetting{
						Domain:               "google.com",
						TitleGoquerySelector: "head>title",
						DatesGoquerySelector: "dates>date",
						FocusIndexFrom:       0,
						FocusIndexTo:         5,
					}, nil)
				rpo.EXPECT().UpdateWebsite(&model.Website{
					UUID:       "uuid",
					URL:        "https://google.com",
					Title:      "title",
					RawContent: "date-1\ndate-2\ndate-3\ndate-4",
					UpdateTime: time.Now().UTC().Truncate(time.Second),
					Conf: &config.WebsiteConfig{
						MaxDateLength: 1,
						Separator:     "\n",
					},
				}).Return(nil)

				return rpo
			},
			cli: MockClient{get: func(s string) (*http.Response, error) {
				return &http.Response{
					Body: io.NopCloser(strings.NewReader(mockRespWithDates)),
				}, nil
			}},
			params: rueidiscompat.XMessage{
				ID: "123",
				Values: map[string]interface{}{
					"task_name": "update_website_task",
					"web":       `{"uuid":"uuid","url":"https://google.com","title":"title","raw_content":"content","update_time":"2022-01-01T00:00:00Z","Conf":{"Separator":"\n","MaxDateLength":1}}`,
				},
			},
			name: "happy flow",
		},
		{
			name: "cancelled context",
			getCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			},
			getRpo: func(ctrl *gomock.Controller) repository.Repostory {
				rpo := mockrepo.NewMockRepostory(ctrl)

				return rpo
			},
			cli: MockClient{get: func(s string) (*http.Response, error) {
				return &http.Response{
					Body: io.NopCloser(strings.NewReader(mockRespWithDates)),
				}, nil
			}},
			params: rueidiscompat.XMessage{
				ID: "123",
				Values: map[string]interface{}{
					"task_name": "update_website_task",
					"web":       `{"uuid":"uuid","url":"https://google.com","title":"title","raw_content":"content","update_time":"2022-01-01T00:00:00Z","Conf":{"Separator":"\n","MaxDateLength":1}}`,
				},
			},
			wantError: context.Canceled,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			// t.Parallel()

			service.SetCli(test.cli)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := test.getCtx()
			task := &UpdateWebsiteTask{
				rpo: test.getRpo(ctrl),
			}

			task.lock.Lock()
			err := task.Execute(ctx, test.params)
			assert.ErrorIs(t, err, test.wantError)
		})
	}
}

func TestUpdateWebsiteTask_Name(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		task *UpdateWebsiteTask
		want string
	}{
		{
			name: "happy flow",
			task: &UpdateWebsiteTask{},
			want: "update_website_task",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.task.Name()
			assert.Equal(t, test.want, got)
		})
	}
}
