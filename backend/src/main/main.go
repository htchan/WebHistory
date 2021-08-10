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
		methodNotSupport(res)
		return
	}
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	groupName := req.Form.Get("groupName")
	if url == "" || !strings.HasPrefix(url, "http") {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "{ \"message\" : \"invalid url\" }")
	}
	website := Website{Url: url, GroupName: groupName, AccessTime: time.Now()}
	//TODO: turn this to serial
	func() {
		website.Update()
		fmt.Println(website.Map())
		website.Save()
	} ()
	fmt.Fprintln(res, "{ \"message\" : \"website <" + website.Title + "> inserted\" }")
}

func listWebsites(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodGet {
		methodNotSupport(res)
		return
	}

	websiteGroups := make([][]map[string]interface{}, 0)

	for _, groupName := range GroupNames() {
		websiteGroup := make([]map[string]interface{}, 0)
		for _, url := range Group2Urls(groupName) {
			websiteGroup = append(websiteGroup, Url2Website(url).Map())
		}
		websiteGroups = append(websiteGroups, websiteGroup)
	}

	responseByte, err := json.Marshal(map[string]interface{} { "websiteGroups": websiteGroups})
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(res, err.Error())
		return
	}

	fmt.Fprintln(res, string(responseByte))


	// websites := make([]map[string]interface{}, 0)
	// for _, url := range Urls() {
	// 	websites = append(websites, Url2Website(url).Map())
	// }
	// responseByte, err := json.Marshal(map[string]interface{} { "websites": websites })
	// if err != nil {
	// 	res.WriteHeader(http.StatusInternalServerError)
	// 	fmt.Fprintln(res, err.Error())
	// 	return
	// }
	// fmt.Fprintln(res, string(responseByte))
}

func refreshWebsite(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodPost {
		methodNotSupport(res)
		return
	}
	req.ParseForm()
	url := req.Form.Get("url")
	website := Url2Website(url)
	website.AccessTime = time.Now()
	website.Save()
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
		methodNotSupport(res)
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
	website.Delete()
}

func changeWebsiteGroup(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodPost {
		methodNotSupport(res)
		return
	}
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	groupName := req.Form.Get("groupName")
	fmt.Println(url, groupName)
	website := Url2Website(url)
	if !isSubSet(website.Title, groupName) || len(website.Title) == 0 {
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
		website := Url2Website(url)
		fmt.Println(website.Url)
		website.Update()
		fmt.Println(website.UpdateTime)
		website.Save()
	}
	fmt.Println(time.Now(), "regular update finish")
}

func main() {
	fmt.Println("hello")
	openDatabase("./database/websites.db")
	fmt.Println(database)
	go func() {
		regularUpdateWebsites()
		for range time.Tick(time.Hour * 23) { regularUpdateWebsites() }
	}()
	http.HandleFunc("/api/web-history/websites/create", createWebsite)
	http.HandleFunc("/api/web-history/list", listWebsites)
	http.HandleFunc("/api/web-history/websites/refresh", refreshWebsite)
	http.HandleFunc("/api/web-history/websites/delete", deleteWebsite)
	http.HandleFunc("/api/web-history/group/change", changeWebsiteGroup)
	log.Fatal(http.ListenAndServe(":9105", nil))
}
