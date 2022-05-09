package website

import (
	"errors"
	"net/http"
	"database/sql"
	"log"
	"fmt"
	"strings"
	"encoding/json"
	"time"
	"github.com/julienschmidt/httprouter"
	"github.com/htchan/WebHistory/internal/utils"
)

func optionsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Headers", "*")
	res.Header().Set("Access-Control-Allow-Methods", "*")
	res.WriteHeader(http.StatusOK)
	return
}

func writeError(res http.ResponseWriter, statusCode int, err error) {
	res.WriteHeader(statusCode)
	fmt.Println(res, fmt.Sprintf(`{ "error": "%v" }`, err))
}

var UnauthorizedError = errors.New("unauthorized")
var InvalidParamsError = errors.New("invalid params")

func UserUUID(req *http.Request) (string, error) {
	token := req.Header.Get("Authorization")
	if token == "" {
		return "", UnauthorizedError
	}
	userUUID := utils.FindUserByToken(token)
	return userUUID, nil
}

func WebsiteParams(req *http.Request) (string, error) {
	err := req.ParseForm()
	if err != nil {
		return "", InvalidParamsError
	}
	url := req.Form.Get("url")
	if url == "" || !strings.HasPrefix(url, "http") {
		return "", InvalidParamsError
	}
	return url, nil
}

func createWebsiteHandler(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Header().Set("Access-Control-Allow-Origin", "*")
		userUUID, err := UserUUID(req)
		if err != nil {
			writeError(res, http.StatusUnauthorized, err)
			return
		}
		url, err := WebsiteParams(req)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		log.Println("create-website.start", fmt.Sprintf(`{ "url": "%v" }`, url))
		w := NewWebsite(url, userUUID)
		err = w.Create(db)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		log.Println("create-website.complete", w.Map())
		fmt.Fprintln(res, fmt.Sprintf(`{ "message": "website <%v> inserted" }`, w.Title))
	}
}

func listWebsiteHandler(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Header().Set("Access-Control-Allow-Origin", "*")
		userUUID, err := UserUUID(req)
		if err != nil {
			writeError(res, http.StatusUnauthorized, err)
			return
		}
		websites, err := FindAllUserWebsites(db, userUUID)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		websiteGroups := WebsitesToWebsiteGroups(websites)
		err = json.NewEncoder(res).Encode(
			map[string][]WebsiteGroup{"website_groups": websiteGroups},
		)
		if err != nil {
			writeError(res, http.StatusInternalServerError, errors.New("parse response error"))
		}
	}
}

func refreshWebsiteHandler(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Header().Set("Access-Control-Allow-Origin", "*")
		userUUID, err := UserUUID(req)
		if err != nil {
			writeError(res, http.StatusUnauthorized, err)
			return
		}
		url, err := WebsiteParams(req)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		w, err := FindUserWebsite(db, userUUID, url)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		w.AccessTime = time.Now()
		err = w.Save(db)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		err = json.NewEncoder(res).Encode(w)
		if err != nil {
			writeError(res, http.StatusInternalServerError, err)
		}
	}
}

func deleteParams(req *http.Request) (string, error) {
	url := req.URL.Query().Get("url")
	if url == "" || !strings.HasPrefix(url, "http") {
		return "", InvalidParamsError
	}
	return url, nil
}

func deleteWebsiteHandler(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Header().Set("Access-Control-Allow-Origin", "*")
		userUUID, err := UserUUID(req)
		if err != nil {
			writeError(res, http.StatusUnauthorized, err)
			return
		}
		url, err := deleteParams(req)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			log.Println("delete-website", err)
			return
		}
		log.Println("delete-website", fmt.Sprintf(`{ "url": %v }"`, url))
		w, err := FindUserWebsite(db, userUUID, url)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			log.Println("delete-website", err)
			return
		}
		err = w.Delete(db)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			log.Println("delete-website", err)
			return
		}
	}
}

func ChangeGroupNameParams(req *http.Request) (url, groupName string, err error) {
	err = req.ParseForm()
	if err != nil {
		return "", "", InvalidParamsError
	}
	url = req.Form.Get("url")
	groupName = req.Form.Get("group_name")
	if url == "" || !strings.HasPrefix(url, "http") || groupName == "" {
		return "", "", InvalidParamsError
	}
	return url, groupName, nil
}

func changeWebsiteGroupHandler(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Header().Set("Access-Control-Allow-Origin", "*")
		userUUID, err := UserUUID(req)
		if err != nil {
			writeError(res, http.StatusUnauthorized, err)
			return
		}
		url, groupName, err := ChangeGroupNameParams(req)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		log.Println("create-website-group.start", fmt.Sprintf(`{ "url": "%v", "group_name": "%v" }`, url, groupName))
		w, err := FindUserWebsite(db, userUUID, url)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		if !utils.IsSubSet(w.Title, strings.ReplaceAll(groupName, " ", "")) || len(w.Title) == 0 {
			writeError(res, http.StatusBadRequest, errors.New("invalid group name"))
			return
		}
		w.GroupName = groupName
		err = w.Save(db)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		err = json.NewEncoder(res).Encode(w)
		if err != nil {
			writeError(res, http.StatusInternalServerError, errors.New("parse response error"))
			return
		}
	}
}


func AddWebsiteRoutes(router *httprouter.Router, db *sql.DB) {
	router.HandlerFunc(http.MethodOptions, "/api/web-history/websites", optionsHandler)
	router.HandlerFunc(http.MethodOptions, "/api/web-history/websites/refresh", optionsHandler)
	router.HandlerFunc(http.MethodOptions, "/api/web-history/websites/change-group", optionsHandler)

	router.HandlerFunc(http.MethodPost, "/api/web-history/websites", createWebsiteHandler(db))
	router.HandlerFunc(http.MethodGet, "/api/web-history/websites", listWebsiteHandler(db))
	router.HandlerFunc(http.MethodPut, "/api/web-history/websites/refresh", refreshWebsiteHandler(db))
	router.HandlerFunc(http.MethodDelete, "/api/web-history/websites", deleteWebsiteHandler(db))
	router.HandlerFunc(http.MethodPut, "/api/web-history/websites/change-group", changeWebsiteGroupHandler(db))
}
//3wb1lxcYgkOe5IN4XebSzyDFwoepnxKbeLIFb26vitqMvvLwyTM7D4Uf5OpmmrXS