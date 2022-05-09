package main

import (
	"fmt"
	"net/http"
	"log"
	"os"

	"github.com/htchan/WebHistory/pkg/website"

	"github.com/htchan/ApiParser"
	"github.com/julienschmidt/httprouter"
)

func main() {
	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	fmt.Println("hello")
	db, err := website.OpenDatabase(os.Getenv("database_volume"))
	if err != nil {
		log.Println("faile to open database", os.Getenv("database_volume"))
		return
	}
	defer db.Close()
	router := httprouter.New()
	website.AddWebsiteRoutes(router, db)

	log.Fatal(http.ListenAndServe(":9105", router))
}
