package main

import (
	"log"
	"sync"
	"time"

	"github.com/htchan/ApiParser"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repo"
	"github.com/htchan/WebHistory/internal/service"
	"github.com/htchan/WebHistory/internal/utils"
)

const INTERVAL = 5

func generateHostChannels(websites []model.Website) chan chan model.Website {
	hostChannels := make(chan chan model.Website)
	hostChannelMap := make(map[string]chan model.Website)
	go func(hostChannels chan chan model.Website) {
		var wg sync.WaitGroup
		for _, web := range websites {
			if web.Host() == "" {
				continue
			}
			wg.Add(1)
			hostChannel, ok := hostChannelMap[web.Host()]
			if !ok {
				newChannel := make(chan model.Website)
				hostChannelMap[web.Host()] = newChannel
				go func(newChannel chan model.Website, web model.Website) {
					defer wg.Done()
					hostChannels <- newChannel
					newChannel <- web
				}(newChannel, web)
			} else {
				go func(web model.Website) {
					defer wg.Done()
					hostChannel <- web
				}(web)
			}
		}
		wg.Wait()
		for key := range hostChannelMap {
			close(hostChannelMap[key])
		}
		close(hostChannels)
	}(hostChannels)
	return hostChannels
}

func regularUpdateWebsites(r repo.Repostory) {
	log.Println("start batch update")
	webs, err := r.FindWebsites()
	if err != nil {
		log.Println("fail to fetch websites from DB:", err)
		return
	}
	var wg sync.WaitGroup
	for hostChannel := range generateHostChannels(webs) {
		wg.Add(1)
		go func(hostChannel chan model.Website) {
			for web := range hostChannel {
				service.Update(r, &web)
				time.Sleep(INTERVAL * time.Second)
			}
			wg.Done()
		}(hostChannel)
	}
	wg.Wait()

	log.Println("complete batch update")
}

func main() {
	err := utils.Migrate()
	if err != nil {
		log.Println("migration failed:", err)
		return
	}
	ApiParser.SetDefault(ApiParser.FromDirectory("/api_parser"))
	db, err := utils.OpenDatabase()
	if err != nil {
		log.Println("open database failed:", err)
		return
	}
	defer db.Close()
	service.AggregateBackup("/backup")
	err = utils.Backup(db)
	if err != nil {
		log.Println("backup database failed:", err)
		return
	}
	r := repo.NewPsqlRepo(db)
	regularUpdateWebsites(r)
}
