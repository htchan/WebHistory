package logging

import (
	"log"
)

func Log(prefix string, data interface{}) {
	log.Printf("web-history.%v %v", prefix, data)
}

func LogUpdate(url string, data interface{}) {
	log.Printf("web-history.update.[%v] %v", url, data)
}

func LogRequest(action string, data map[string]interface{}) {
	log.Printf("web-history.request.%v.%v", action, data)
}

func LogBatchStatus(status string) {
	log.Printf("web-history.batch.%v", status)
}