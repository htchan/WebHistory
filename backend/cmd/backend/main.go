package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/htchan/WebHistory/internal/repo"
	"github.com/htchan/WebHistory/internal/router/website"
	"github.com/htchan/WebHistory/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/htchan/ApiParser"
)

func main() {
	err := utils.Migrate()
	if err != nil {
		log.Println("failed to migrate", err)
		return
	}
	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	db, err := utils.OpenDatabase()
	if err != nil {
		log.Println("failed to open database", err)
		return
	}
	defer db.Close()
	rpo := repo.NewPsqlRepo(db)
	r := chi.NewRouter()
	website.AddRoutes(r, rpo)

	server := http.Server{
		Addr:         ":9105",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
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
