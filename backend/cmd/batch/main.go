package main

import (
	"os"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/htchan/WebHistory/pkg/website"
	"github.com/htchan/WebHistory/internal/utils"

	"github.com/htchan/ApiParser"
)

func generateHostChannels(websites []website.Website) chan chan website.Website {
	hostChannels := make(chan chan website.Website)
	hostChannelMap := make(map[string]chan website.Website)
	go func(hostChannels chan chan website.Website) {
		var wg sync.WaitGroup
		for _, web := range websites {
			if web.Host() == "" {
				continue
			}
			wg.Add(1)
			hostChannel, ok := hostChannelMap[web.Host()]
			if !ok {
				newChannel := make(chan website.Website)
				hostChannelMap[web.Host()] = newChannel
				go func(newChannel chan website.Website, web website.Website) {
					defer wg.Done()
					hostChannels <- newChannel
					newChannel <- web
				} (newChannel, web)
			} else {
				go func(web website.Website) {
					defer wg.Done()
					hostChannel <- web
				} (web)
			}
		}
		wg.Wait()
		for key := range hostChannelMap {
			close(hostChannelMap[key])
		}
		close(hostChannels)
	} (hostChannels)
	return hostChannels
}

func regularUpdateWebsites(db *sql.DB) {
	log.Println("start")
	websites, err := website.FindAllWebsites(db)
	if err != nil {
		log.Println("fail to fetch websites in DB:", err)
		return
	}
	var wg sync.WaitGroup
	for hostChannel := range generateHostChannels(websites) {
		wg.Add(1)
		go func(hostChannel chan website.Website) {
			for web := range hostChannel {
				log.Println(web.URL, "start", nil)
				web.Update()
				log.Println(web.URL, "info", web.Map())
				web.Save(db)
				log.Println(web.URL, "finish", nil)
				time.Sleep(1 * time.Second)
			}
			wg.Done()
		}(hostChannel)
	}
	wg.Wait()

	log.Println("complete")
}

func main() {
	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	db, err := utils.OpenDatabase(os.Getenv("database_volume"))
	if err != nil {
		log.Println("open database failed:", err)
	}
	defer db.Close()
	regularUpdateWebsites(db)		
}
