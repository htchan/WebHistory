package main

import (
	"os"
	"time"

	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/executor"
	"github.com/htchan/WebHistory/internal/jobs/batchupdate"
	"github.com/htchan/WebHistory/internal/jobs/websiteupdate"
	"github.com/htchan/WebHistory/internal/repository/sqlc"
	"github.com/htchan/WebHistory/internal/shutdown"
	"github.com/htchan/WebHistory/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

	if err = utils.Migrate(&conf.DatabaseConfig); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate")
	}

	shutdownHandler := shutdown.New()

	db, err := utils.OpenDatabase(&conf.DatabaseConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	rpo := sqlc.NewRepo(db, &conf.WebsiteConfig)

	// TODO: use config to define worker number
	exec := executor.NewExecutor(conf.BinConfig.WorkerExecutorCount)

	// start website update job
	websiteUpdateScheduler := websiteupdate.Setup(rpo, conf.BinConfig.WebsiteUpdateSleepInterval)
	exec.Register(websiteUpdateScheduler.Publisher())
	go websiteUpdateScheduler.Start()

	// start batch update job
	batchUpdateScheduer := batchupdate.Setup(rpo, websiteUpdateScheduler)
	exec.Register(batchUpdateScheduer.Publisher())
	go batchUpdateScheduer.Start()

	shutdownHandler.Register("batchupdate.Scheduler", batchUpdateScheduer.Stop)
	shutdownHandler.Register("websiteupdate.Scheduler", websiteUpdateScheduler.Stop)
	shutdownHandler.Register("executor", exec.Stop)
	shutdownHandler.Register("database", db.Close)

	go exec.Start()

	shutdownHandler.Listen(5 * time.Second)
}
