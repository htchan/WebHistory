package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/htchan/WebHistory/internal/repository/sqlc"
	"github.com/htchan/WebHistory/internal/service"
	"github.com/htchan/WebHistory/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func websiteChannelGroupedByHost(websites []model.Website) map[string]chan model.Website {
	hostWebsitesMap := make(map[string][]model.Website)

	// group websites by Host, channel blocking will only affect its iteration
	for _, website := range websites {
		host := website.Host()
		hostWebsitesMap[host] = append(hostWebsitesMap[host], website)
	}

	hostChannelMap := make(map[string]chan model.Website)

	for host, groupedWebsites := range hostWebsitesMap {
		host := host
		groupedWebsites := groupedWebsites
		hostChannelMap[host] = make(chan model.Website)

		go func(host string, websites []model.Website) {
			for _, website := range websites {
				hostChannelMap[host] <- website
			}
			close(hostChannelMap[host])
		}(host, groupedWebsites)
	}

	return hostChannelMap
}

func regularUpdateWebsites(r repository.Repostory, conf *config.BatchBinConfig) {
	tr := otel.Tracer("process")
	ctx, span := tr.Start(context.Background(), "batch")
	defer span.End()

	webs, err := r.FindWebsites()
	if err != nil {
		log.Error().Err(err).Msg("fail to fetch websites from DB")
		return
	}

	var wg sync.WaitGroup
	for host, channel := range websiteChannelGroupedByHost(webs) {
		wg.Add(1)

		go func(host string, websites chan model.Website) {
			ctx, span := tr.Start(ctx, host)
			defer span.End()

			for web := range websites {
				ctx := log.With().
					Str("job_uuid", uuid.Must(uuid.NewUUID()).String()).
					Logger().
					WithContext(ctx)
				service.Update(ctx, r, &web)
				time.Sleep(conf.SleepInterval)
			}
			wg.Done()
		}(host, channel)
	}

	wg.Wait()
}

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

func closeTracer(tp *tracesdk.TracerProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
}

func main() {
	outputPath := os.Getenv("OUTPUT_PATH")
	if outputPath != "" {
		writer, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err == nil {
			log.Logger = log.Logger.Output(writer)
		}
	}

	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.99999Z07:00"

	memStart := findMemUsage()
	printMemUsage(memStart)

	conf, err := config.LoadBatchConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("load config failed")
	}

	tp, err := tracerProvider(conf.TraceConfig)
	if err != nil {
		log.Error().Err(err).Msg("init tracer failed")
	}
	defer closeTracer(tp)

	if err = utils.Migrate(&conf.DatabaseConfig); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

	db, err := utils.OpenDatabase(&conf.DatabaseConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("open database failed")
	}
	defer db.Close()

	// service.AggregateBackup(conf.BackupDirectory)

	r := sqlc.NewRepo(db, &conf.WebsiteConfig)

	regularUpdateWebsites(r, &conf.BinConfig)
	printMemDiff(memStart, findMemUsage())
}
