package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/htchan/ApiParser"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repo"
	"github.com/htchan/WebHistory/internal/service"
	"github.com/htchan/WebHistory/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	INTERVAL     = 5
	TRACE_URL    = "http://monitoring:14268/api/traces"
	SERVICE_NAME = "web-watcher"
)

func generateHostChannels(websites []model.Website) chan chan model.Website {
	hostChannels := make(chan chan model.Website)
	hostChannelMap := make(map[string]chan model.Website)

	go func(hostChannels chan chan model.Website) {
		var wg sync.WaitGroup
		for _, web := range websites {
			if web.Host() == "" {
				continue
			}

			wg.Add(1)
			hostChannel, ok := hostChannelMap[web.Host()]
			if !ok {
				newChannel := make(chan model.Website)
				hostChannelMap[web.Host()] = newChannel
				go func(newChannel chan model.Website, web model.Website) {
					defer wg.Done()
					hostChannels <- newChannel
					newChannel <- web
				}(newChannel, web)
			} else {
				go func(web model.Website) {
					defer wg.Done()
					hostChannel <- web
				}(web)
			}
		}

		wg.Wait()
		for key := range hostChannelMap {
			close(hostChannelMap[key])
		}
		close(hostChannels)
	}(hostChannels)

	return hostChannels
}

func loadSettings(r repo.Repostory) {
	settings, err := r.FindWebsiteSettings()
	if err != nil {
		return
	}

	for _, setting := range settings {
		ApiParser.AddFormatSet(ApiParser.NewFormatSet(
			setting.Domain,
			setting.ContentRegex,
			setting.TitleRegex,
		))
	}
}

func regularUpdateWebsites(r repo.Repostory) {
	tr := otel.Tracer("process")
	ctx, span := tr.Start(context.Background(), "batch")
	defer span.End()

	webs, err := r.FindWebsites()
	if err != nil {
		log.Println("fail to fetch websites from DB:", err)
		return
	}

	var wg sync.WaitGroup
	for hostChannel := range generateHostChannels(webs) {
		wg.Add(1)

		go func(hostChannel chan model.Website) {
			for web := range hostChannel {
				service.Update(ctx, r, &web)
				time.Sleep(INTERVAL * time.Second)
			}
			wg.Done()
		}(hostChannel)
	}

	wg.Wait()
}

func tracerProvider() (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(TRACE_URL)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(SERVICE_NAME),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

func closeTracer(tp *tracesdk.TracerProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Println("shutdown error", err)
	}
}

func main() {
	tp, err := tracerProvider()
	if err != nil {
		log.Println("tracer error", err)
	}
	defer closeTracer(tp)

	err = utils.Migrate()
	if err != nil {
		log.Fatalln("migration failed:", err)
	}

	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	db, err := utils.OpenDatabase()
	if err != nil {
		log.Fatalln("open database failed:", err)
	}
	defer db.Close()

	service.AggregateBackup("/backup")
	err = utils.Backup(db)
	if err != nil {
		log.Fatalln("backup database failed:", err)
	}

	r := repo.NewPsqlRepo(db)

	loadSettings(r)

	regularUpdateWebsites(r)
}
