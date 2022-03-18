package main

import (
	"time"
	"fmt"
	"os"

	"github.com/htchan/WebHistory/pkg/websites"
)

func regularUpdateWebsites() {
	fmt.Println(time.Now(), "regular update")
	for _, website := range websites.FindAllWebsites() {
		fmt.Println(website.Url)
		website.Update()
		fmt.Println(website.UpdateTime)
		website.Save()
	}
	fmt.Println(time.Now(), "regular update finish")
}

func main() {
	websites.OpenDatabase(os.Getenv("database_volume"))
	
	regularUpdateWebsites()		
}
