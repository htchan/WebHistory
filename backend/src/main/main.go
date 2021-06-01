package main

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"log"
)

func add(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method == "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(res, "{ \"error\" : \"method not support\" }")
		return
	}
	req.ParseForm()
	url := req.Form.Get("url")
	web := Website{Url: url}
	web.Update()
	fmt.Fprintln(res, "{ \"message\" : \"website <" + url + "> add success\" }")
}

func list(res http.ResponseWriter, req *http.Request) {
	result := make([]map[string]interface{}, 0)
	for _, url := range Urls() {
		result = append(result, GetWeb(url).Map())
	}
	var responseByte []byte
	err := json.Unmarshal(responseByte, map[string]interface{} { "websites": result })
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		return
	}
	fmt.Fprintln(res, string(responseByte))
}

func regularUpdate() {
	for range time.Tick(time.Hour * 23) {
		for _, url := range Urls() {
			web := GetWeb(url)
			web.Update()
		}
	}
}

func main() {
	fmt.Println("hello")
	openDatabase("/database")
	go regularUpdate()
	http.HandleFunc("/api/web-history/add", add)
	http.HandleFunc("/api/web/history/list", list)
	log.Fatal(http.ListenAndServe(":9105", nil))
}