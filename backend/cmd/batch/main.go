package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/htchan/ApiParser"
	"github.com/htchan/WebHistory/internal/config"
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

func regularUpdateWebsites(r repo.Repostory, conf config.BatchConfig) {
	tr := otel.Tracer("process")
	ctx, span := tr.Start(context.Background(), "batch")
	defer span.End()

	webs, err := r.FindWebsites()
	if err != nil {
		log.Println("fail to fetch websites from DB:", err)
		return
	}

	var wg sync.WaitGroup
	for host, channel := range websiteChannelGroupedByHost(webs) {
		wg.Add(1)

		go func(host string, website chan model.Website) {
			ctx, span := tr.Start(ctx, host)
			defer span.End()

			for web := range website {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Println("shutdown error", err)
	}
}

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("load config failed:", err)
		return
	}

	tp, err := tracerProvider(conf.TraceConfig)
	if err != nil {
		log.Println("tracer error", err)
	}
	defer closeTracer(tp)

	if err = utils.Migrate(&conf.DatabaseConfig); err != nil {
		log.Fatalln("migration failed:", err)
	}

	ApiParser.SetDefault(ApiParser.FromDirectory(conf.ApiParserDirectory))
	db, err := utils.OpenDatabase(&conf.DatabaseConfig)
	if err != nil {
		log.Fatalln("open database failed:", err)
	}
	defer db.Close()

	service.AggregateBackup(conf.BackupDirectory)
	err = utils.Backup(db)
	if err != nil {
		log.Fatalln("backup database failed:", err)
	}

	r := repo.NewPsqlRepo(db)

	loadSettings(r)

	regularUpdateWebsites(r, conf.BatchConfig)
}
