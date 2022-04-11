package main

import (
	"os"

	"github.com/htchan/WebHistory/internal/logging"
	"github.com/htchan/WebHistory/pkg/websites"

	"github.com/htchan/ApiParser"
)

func regularUpdateWebsites() {
	logging.LogBatchStatus("start")
	for _, website := range websites.FindAllWebsites() {
		logging.LogUpdate(website.Url, "start", nil)
		website.Update()
		logging.LogUpdate(website.Url, "info", website.Map())
		website.Save()
		logging.LogUpdate(website.Url, "finish", nil)
	}
	logging.LogBatchStatus("complete")
}

func main() {
	ApiParser.Setup("/api_parser")
	websites.OpenDatabase(os.Getenv("database_volume"))
	
	regularUpdateWebsites()		
}
