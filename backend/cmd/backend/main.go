package main

import (
	"fmt"
	"net/http"
	"log"
	"os"

	"github.com/htchan/WebHistory/pkg/router"
	"github.com/htchan/WebHistory/internal/utils"

	"github.com/htchan/ApiParser"
	"github.com/go-chi/chi/v5"
)

func main() {
	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	fmt.Println("hello")
	db, err := utils.OpenDatabase(os.Getenv("database_volume"))
	if err != nil {
		log.Println("faile to open database", os.Getenv("database_volume"))
		return
	}
	defer db.Close()
	r := chi.NewRouter()
	router.AddWebsiteRoutes(r, db)

	log.Fatal(http.ListenAndServe(":9105", r))
}
