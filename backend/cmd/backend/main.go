package main

import (
	"log"
	"net/http"

	"github.com/htchan/WebHistory/internal/repo"
	"github.com/htchan/WebHistory/internal/router"
	"github.com/htchan/WebHistory/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/htchan/ApiParser"
)

func main() {
	err := utils.Migrate()
	if err != nil {
		panic(err)
	}
	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	db, err := utils.OpenDatabase()
	if err != nil {
		log.Println("faile to open database", err)
		return
	}
	defer db.Close()
	rpo := repo.NewPsqlRepo(db)
	r := chi.NewRouter()
	router.AddWebsiteRoutes(r, rpo)

	log.Fatal(http.ListenAndServe(":9105", r))
}
