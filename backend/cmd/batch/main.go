package main

import (
	"os"

	"github.com/htchan/WebHistory/internal/logging"
	"github.com/htchan/WebHistory/pkg/websites"
)

func regularUpdateWebsites() {
	logging.LogBatchStatus("start")
	for _, website := range websites.FindAllWebsites() {
		logging.LogUpdate(website.Url, "start")
		website.Update()
		logging.LogUpdate(website.Url, website.Map())
		website.Save()
		logging.LogUpdate(website.Url, "finish")
	}
	logging.LogBatchStatus("complete")
}

func main() {
	websites.OpenDatabase(os.Getenv("database_volume"))
	
	regularUpdateWebsites()		
}
