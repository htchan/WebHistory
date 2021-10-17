package main

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"os"

	"github.com/htchan/WebHistory/internal"
	"github.com/htchan/WebHistory/pkg/websites"
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
	website := websites.Website{Url: url, GroupName: groupName, AccessTime: time.Now()}
	website.Update()
	fmt.Println(website.Map())
	website.Save()
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

	for _, groupName := range websites.GroupNames() {
		websiteGroup := make([]map[string]interface{}, 0)
		for _, url := range websites.Group2Urls(groupName) {
			websiteGroup = append(websiteGroup, websites.Url2Website(url).Map())
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
	website := websites.Url2Website(url)
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
	website := websites.Url2Website(url)
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
	website := websites.Url2Website(url)
	if !internal.IsSubSet(website.Title, strings.ReplaceAll(groupName, " ", "")) || len(website.Title) == 0 {
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
	for _, url := range websites.Urls() {
		website := websites.Url2Website(url)
		fmt.Println(website.Url)
		website.Update()
		fmt.Println(website.UpdateTime)
		website.Save()
	}
	fmt.Println(time.Now(), "regular update finish")
}

func main() {
	fmt.Println("hello")
	websites.OpenDatabase(os.Getenv("database_volume"))
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
