package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/repository/sqlc"
	"github.com/htchan/WebHistory/internal/router/website"
	"github.com/htchan/WebHistory/internal/utils"

	"github.com/go-chi/chi/v5"
)

func main() {
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.99999Z07:00"

	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("load config failed")
	}

	if err = utils.Migrate(&conf.DatabaseConfig); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate")
	}

	db, err := utils.OpenDatabase(&conf.DatabaseConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	defer db.Close()

	rpo := sqlc.NewRepo(db, conf)
	r := chi.NewRouter()
	website.AddRoutes(r, rpo, conf)

	server := http.Server{
		Addr:         conf.APIConfig.Addr,
		ReadTimeout:  conf.APIConfig.ReadTimeout,
		WriteTimeout: conf.APIConfig.WriteTimeout,
		IdleTimeout:  conf.APIConfig.IdleTimeout,
		Handler:      r,
	}

	go func() {
		log.Debug().Msg("start http server")

		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("backend stopped")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	log.Debug().Msg("received interrupt signal")

	// Setup graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
