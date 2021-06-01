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
	if req.Method != "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(res, "{ \"error\" : \"method not support\" }")
		return
	}
	req.ParseForm()
	url := req.Form.Get("url")
	web := Website{Url: url, AccessTime: time.Now()}
	web.Update()
	web.Save()
	fmt.Println(web.Response())
	fmt.Fprintln(res, "{ \"message\" : \"website <" + url + "> add success\" }")
}

func list(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	result := make([]map[string]interface{}, 0)
	fmt.Println(Urls())
	for _, url := range Urls() {
		result = append(result, GetWeb(url).Response())
	}
	fmt.Println()
	responseByte, err := json.Marshal(map[string]interface{} { "websites": result })
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		return
	}
	fmt.Fprintln(res, string(responseByte))
}

func refresh(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(res, "{ \"error\" : \"method not support\" }")
		return
	}
	req.ParseForm()
	url := req.Form.Get("url")
	web := GetWeb(url)
	web.AccessTime = time.Now()
	fmt.Println("web")
	web.Save()
}

func regularUpdate() {
	for range time.Tick(time.Hour * 23) {
		for _, url := range Urls() {
			web := GetWeb(url)
			web.Update()
			web.Save()
		}
	}
}

func main() {
	fmt.Println("hello")
	openDatabase("./database/websites.db")
	fmt.Println(database)
	go regularUpdate()
	http.HandleFunc("/api/web-history/add", add)
	http.HandleFunc("/api/web-history/list", list)
	http.HandleFunc("/api/web-history/refresh", refresh)
	log.Fatal(http.ListenAndServe(":9105", nil))
}