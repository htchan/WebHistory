package main

import (
	"os"
	"database/sql"
	"log"
	"sync"

	"github.com/htchan/WebHistory/pkg/website"

	"github.com/htchan/ApiParser"
)

func generateHostChannels(websites []website.Website) chan chan website.Website {
	hostChannels := make(chan chan website.Website)
	hostChannelMap := make(map[string]chan website.Website)
	go func(hostChannels chan chan website.Website) {
		for _, w := range websites {
			if w.Host() == "" {
				continue
			}
			hostChannel, ok := hostChannelMap[w.Host()]
			if !ok {
				newChannel := make(chan website.Website)
				hostChannelMap[w.Host()] = newChannel
				hostChannels <- newChannel
				newChannel <- w
			} else {
				hostChannel <- w
			}
		}
		for key := range hostChannelMap {
			close(hostChannelMap[key])
		}
		close(hostChannels)
	}(hostChannels)
	return hostChannels
}

func regularUpdateWebsites(db *sql.DB) {
	log.Println("start")
	websites, err := website.FindAllWebsites(db)
	if err != nil {
		log.Println("fail to fetch websites:", err)
		return
	}
	var wg sync.WaitGroup
	for hostChannel := range generateHostChannels(websites) {
		go func(hostChannel chan website.Website) {
			wg.Add(1)
			for w := range hostChannel {
				log.Println(w.URL, "start", nil)
				w.Update()
				log.Println(w.URL, "info", w.Map())
				w.Save(db)
				log.Println(w.URL, "finish", nil)
			}
			wg.Done()
		}(hostChannel)
	}
	wg.Wait()

	log.Println("complete")
}

func main() {
	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	db, err := website.OpenDatabase(os.Getenv("database_volume"))
	if err != nil {
		log.Println("open database failed:", err)
	}
	defer db.Close()
	regularUpdateWebsites(db)		
}
