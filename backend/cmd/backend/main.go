package main

import (
	"log"
	"net/http"

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

	log.Println("start http server")
	log.Fatal(http.ListenAndServe(":9105", r))
}
