package main

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/htchan/WebHistory/internal/repository/sqlc"
	"github.com/htchan/WebHistory/internal/utils"
	updatewebsite "github.com/htchan/WebHistory/internal/workers/update_website"
	shutdown "github.com/htchan/goshutdown"
	"github.com/htchan/goworkers"
	"github.com/htchan/goworkers/stream/redis"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// TODO: move tracer to helper
func tracerProvider(conf config.TraceConfig) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(conf.TraceURL)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(conf.TraceServiceName),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func main() {
	outputPath := os.Getenv("OUTPUT_PATH")
	if outputPath != "" {
		writer, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err == nil {
			log.Logger = log.Logger.Output(writer)
			defer writer.Close()
		} else {
			log.Fatal().
				Err(err).
				Str("output_path", outputPath).
				Msg("set logger output failed")
		}
	}

	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.99999Z07:00"

	conf, err := config.LoadWorkerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("load config failed")
	}

	tp, err := tracerProvider(conf.TraceConfig)
	if err != nil {
		log.Error().Err(err).Msg("init tracer failed")
	}

	if err = utils.Migrate(&conf.DatabaseConfig); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate")
	}

	shutdown.LogEnabled = true
	shutdownHandler := shutdown.New(syscall.SIGINT, syscall.SIGTERM)

	db, err := utils.OpenDatabase(&conf.DatabaseConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	rpo := sqlc.NewRepo(db, &conf.WebsiteConfig)

	ctx := context.Background()

	shutdownHandler.Register("database", db.Close)
	shutdownHandler.Register("tracer", func() error {
		return tp.Shutdown(ctx)
	})

	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{conf.RedisURL},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to redis")

		return
	}

	tasks := make(map[string]goworkers.Task)

	for _, host := range conf.BinConfig.SupportHosts {
		tasks[host] = updatewebsite.NewUpdateWebsiteTask(
			redis.NewRedisStream(
				redisClient,
				fmt.Sprintf("web-history/update-books/%s", host),
				"web-history/subscriber",
				"web-history/subscriber",
				100*time.Millisecond,
			),
			rpo,
			conf.BinConfig.WebsiteUpdateSleepInterval,
		)
	}

	deployTask(ctx, conf, rpo, tasks)

	shutdownHandler.Listen(60 * time.Second)
}

func deployTask(ctx context.Context, conf *config.WorkerConfig, rpo repository.Repostory, tasks map[string]goworkers.Task) {
	log.Info().Msg("start publish tasks")

	webs, err := rpo.FindWebsites()
	if err != nil {
		log.Error().Err(err).Msg("failed to list websites")

		return
	}

	fmt.Println(webs)

	for _, web := range webs {
		web := web
		task, ok := tasks[web.Host()]
		if !ok {
			log.Warn().Str("host", web.Host()).Msg("undefined tasks")
			continue
		}

		err := task.Publish(ctx, web)
		if err != nil {
			log.Error().Err(err).Str("website", web.URL).Msg("failed to publish task")

			continue
		}
		log.Debug().Str("website", web.URL).Msg("publish task")
	}

	log.Info().Msg("publish tasks finished")
}
