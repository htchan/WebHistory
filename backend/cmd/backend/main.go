package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/repo"
	"github.com/htchan/WebHistory/internal/router/website"
	"github.com/htchan/WebHistory/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/htchan/ApiParser"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("load config failed:", err)
		return
	}

	if err = utils.Migrate(&conf.DatabaseConfig); err != nil {
		log.Println("failed to migrate", err)
		return
	}
	ApiParser.SetDefault(ApiParser.FromDirectory(conf.ApiParserDirectory))
	db, err := utils.OpenDatabase(&conf.DatabaseConfig)
	if err != nil {
		log.Println("failed to open database", err)
		return
	}
	defer db.Close()
	rpo := repo.NewPsqlRepo(db, conf)
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
		log.Println("start http server")

		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("backend stopped: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	log.Println("received interrupt signal")

	// Setup graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
