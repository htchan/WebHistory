package main

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"os"

	"github.com/htchan/WebHistory/internal/utils"
	"github.com/htchan/WebHistory/internal/logging"
	"github.com/htchan/WebHistory/pkg/websites"

	"github.com/htchan/ApiParser"
)

func methodNotSupport(res http.ResponseWriter) {
	res.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(res, "{ \"error\" : \"method not support\" }")
}

func unauthorized(res http.ResponseWriter) {
	res.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintln(res, "{ \"error\" : \"unauthorized\" }")
}

func userServiceLogin(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")

	if req.Method != http.MethodPost {
		methodNotSupport(res)
		return
	}

	err := req.ParseForm()
	if err != nil {panic(err)}
	token := req.Form.Get("token")
	userUUID := utils.FindUserByToken(token)

	if userUUID == "" {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		unauthorized(res)
		return
	}
	http.Redirect(res, req, os.Getenv("WEB_HISTORY_FRONTEND_TOKEN_URL") + "?token=" + token, 302)
}

func createWebsite(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodPost {
		methodNotSupport(res)
		return
	}
	token := req.Header.Get("Authorization")
	if token == "" {
		unauthorized(res)
		return
	}
	userUUID := utils.FindUserByToken(token)
	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	logging.LogRequest("create-website.start", map[string]interface{} { "url": url })
	groupName := req.Form.Get("groupName")
	if url == "" || !strings.HasPrefix(url, "http") {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, "{ \"message\" : \"invalid url\" }")
	}
	website := websites.Website{UserUUID: userUUID, Url: url, GroupName: groupName, AccessTime: time.Now()}
	website.Update()
	website.Save()
	logging.LogRequest("create-website.complete", website.Map())
	fmt.Fprintln(res, "{ \"message\" : \"website <" + website.Title + "> inserted\" }")
}

func listWebsites(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	if req.Method != http.MethodGet {
		methodNotSupport(res)
		return
	}
	token := req.Header.Get("Authorization")
	if token == "" {
		unauthorized(res)
		return
	}
	userUUID := utils.FindUserByToken(token)

	websiteGroups := make([][]map[string]interface{}, 0)

	for _, groupName := range websites.FindAllGroupNames(userUUID) {
		websiteGroup := make([]map[string]interface{}, 0)
		for _, url := range websites.FindUrlsByGroupName(userUUID, groupName) {
			websiteGroup = append(websiteGroup, websites.FindWebsiteByUrl(userUUID, url).Map())
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
	token := req.Header.Get("Authorization")
	if token == "" {
		unauthorized(res)
		return
	}
	userUUID := utils.FindUserByToken(token)

	req.ParseForm()
	url := req.Form.Get("url")
	logging.LogRequest("refresh-website", map[string]interface{} { "url": url })
	website := websites.FindWebsiteByUrl(userUUID, url)
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
	token := req.Header.Get("Authorization")
	if token == "" {
		unauthorized(res)
		return
	}
	userUUID := utils.FindUserByToken(token)

	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	logging.LogRequest("delete-website", map[string]interface{} { "url": url })
	website := websites.FindWebsiteByUrl(userUUID, url)
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
	token := req.Header.Get("Authorization")
	if token == "" {
		unauthorized(res)
		return
	}
	userUUID := utils.FindUserByToken(token)

	err := req.ParseForm()
	if err != nil {panic(err)}
	url := req.Form.Get("url")
	groupName := req.Form.Get("groupName")
	logging.LogRequest("change-website-group", map[string]interface{} {"url": url, "group": groupName})
	website := websites.FindWebsiteByUrl(userUUID, url)
	if !utils.IsSubSet(website.Title, strings.ReplaceAll(groupName, " ", "")) || len(website.Title) == 0 {
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

func main() {
	ApiParser.Setup("/api_parser")
	fmt.Println("hello")
	websites.OpenDatabase(os.Getenv("database_volume"))
	
	http.HandleFunc("/api/web-history/user_service/login", userServiceLogin)
	http.HandleFunc("/api/web-history/websites/create", createWebsite)
	http.HandleFunc("/api/web-history/list", listWebsites)
	http.HandleFunc("/api/web-history/websites/refresh", refreshWebsite)
	http.HandleFunc("/api/web-history/websites/delete", deleteWebsite)
	http.HandleFunc("/api/web-history/group/change", changeWebsiteGroup)
	log.Fatal(http.ListenAndServe(":9105", nil))
}
