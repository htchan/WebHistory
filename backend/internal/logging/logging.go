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

func LogUpdate(url string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("web-history.update.[%v] %v", url, data)
	} else {
		log.Printf("web-history.update.[%v] %v", url, string(jsonData))
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