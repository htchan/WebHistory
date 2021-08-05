package main

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"log"
	"strings"
)

func methodNotSupport(res http.ResponseWriter) {
	res.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(res, "{ \"error\" : \"method not support\" }")
}

func createWebsite(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodPost {
		methodNotSupport()
		return
	}
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	if url == "" || !strings.HasPrefix(url, "http") {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "{ \"error\" : \"invalid url\" }")
	}
	website := Website{Url: url, AccessTime: time.Now()}
	//TODO: turn this to serial
	go func() {
		website.Update()
		fmt.Println(website.Map())
		website.Save()
	} ()
	fmt.Fprintln(res, "{ \"message\" : \"website <" + url + "> is put into create queue\" }")
}

func listWebistes(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodGet {
		methodNotSupport()
		return
	}
	websites := make([]map[string]interface{}, 0)
	for _, url := range Urls() {
		websites = append(websites, Url2Website(url).Map())
	}
	responseByte, err := json.Marshal(map[string]interface{} { "websites": websites })
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		return
	}
	fmt.Fprintln(res, string(responseByte))
}

func refreshWebsite(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodPost {
		methodNotSupport()
		return
	}
	req.ParseForm()
	url := req.Form.Get("url")
	website := Url2Website(url)
	website.AccessTime = time.Now()
	web.Save()
	responseByte, err := json.Marshal(map[string]interface{} { "website": website.Map() })
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		return
	}
	fmt.Fprintln(res, string(responseByte))
}

func deleteWebsite(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodPost {
		methodNotSupport()
		return
	}
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	fmt.Println(url)
	website := Url2Website(url)
	if website.UpdateTime.Unix() == -62135596800 && website.AccessTime.Unix() == -62135596800 {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "{ \"error\" : \"website not found\" }")
		return
	}
	web.Delete()
}

func changeWebsiteGroup(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodPost {
		methodNotSupport()
		return
	}
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	groupName := req.Form.Get("groupName")
	fmt.Println(url, groupName)
	website := Url2Website(url)
	if isSubSet(website.Title, groupName) {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "{ \"error\" : \"invalid group name\" }")
		return
	}
	if website.UpdateTime.Unix() == -62135596800 && website.AccessTime.Unix() == -62135596800 {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "{ \"error\" : \"website not found\" }")
		return
	}
	website.GroupName = groupName
	website.Save()
	responseByte, err := json.Marshal(map[string]interface{} { "website": website.Map() })
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		return
	}
	fmt.Fprintln(res, string(responseByte))
}

func regularUpdateWebsites() {
	fmt.Println(time.Now(), "regular update")
	for _, url := range Urls() {
		fmt.Println(url)
		web := Url2Website(url)
		web.Update()
		web.Save()
	}
	fmt.Println(time.Now(), "regular update finish")
}

func main() {
	fmt.Println("hello")
	openDatabase("./database/websites.db")
	fmt.Println(database)
	go func() {
		regularUpdate()
		for range time.Tick(time.Hour * 23) { regularUpdateWebsites() }
	}()
	http.HandleFunc("/api/web-history/create", createWebsite)
	http.HandleFunc("/api/web-history/list", listWebsites)
	http.HandleFunc("/api/web-history/refresh", refreshWebsite)
	http.HandleFunc("/api/web-history/delete", deleteWebsite)
	http.HandleFunc("/api/web-history/group", changeWebsiteGroup)
	log.Fatal(http.ListenAndServe(":9105", nil))
}
