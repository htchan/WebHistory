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
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	if url == nil || url == "" || !strings.HasPrefix(url, "http") {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Println(res, "{ \"error\" : \"invalid url format\" }")
	}
	web := Website{Url: url, AccessTime: time.Now()}
	go func() {
		web.Update()
		fmt.Println(web.Response())
		web.Save()
	} ()
	fmt.Fprintln(res, "{ \"message\" : \"website <" + url + "> add is put into queue\" }")
}

func list(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	result := make([]map[string]interface{}, 0)
	for _, url := range Urls() {
		result = append(result, GetWeb(url).Response())
	}
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
	web.Save()
}

func delete(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(res, "{ \"error\" : \"method not support\" }")
		return
	}
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	fmt.Println(url)
	web := GetWeb(url)
	if web.UpdateTime.Unix() == -62135596800 && web.AccessTime.Unix() == -62135596800 {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "{ \"error\" : \"website not found\" }")
		return
	}
	web.Delete()
	fmt.Fprintln(res, "{ \"error\" : \"website not found\" }")
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
	http.HandleFunc("/api/web-history/delete", delete)
	log.Fatal(http.ListenAndServe(":9105", nil))
}