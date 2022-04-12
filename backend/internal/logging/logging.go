package logging

import (
	"log"
	"encoding/json"
)

func Log(prefix string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("web-history.%v %v", prefix, data)
	} else {
		log.Printf("web-history.%v %v", prefix, string(jsonData))
	}
}

func LogUpdate(url string, status string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("web-history.update.%v.[%v] %v", status, url, data)
	} else {
		log.Printf("web-history.update.%v.[%v] %v", status, url, string(jsonData))
	}
}

func LogRequest(action string, data map[string]interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("web-history.request.%v.%v", action, data)
	} else {
		log.Printf("web-history.request.%v.%v", action, string(jsonData))
	}
}

func LogBatchStatus(status string) {
	log.Printf("web-history.batch.%v", status)
}
