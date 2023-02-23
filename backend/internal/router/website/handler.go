package website

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/htchan/WebHistory/internal/service"
)

func getAllWebsiteGroupsHandler(r repository.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		webs, err := r.FindUserWebsites(userUUID)
		if err != nil {
			log.Printf("path: %s; user_uuid: %s; error: %s", req.URL, userUUID, err)
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
			return
		}
		log.Printf("path: %s; user_uuid: %s", req.URL, userUUID)

		json.NewEncoder(res).Encode(map[string]interface{}{
			"website_groups": webs.WebsiteGroups(),
		})
	}
}

func getWebsiteGroupHandler(r repository.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		groupName := chi.URLParam(req, "groupName")
		webs, err := r.FindUserWebsitesByGroup(userUUID, groupName)
		if err != nil || len(webs) == 0 {
			log.Printf("path: %s; user_uuid: %s; group_name: %s; error: %s", req.URL, userUUID, groupName, err)
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
			return
		}
		log.Printf("path: %s; user_uuid: %s; group_name: %s", req.URL, userUUID, groupName)

		json.NewEncoder(res).Encode(
			map[string]interface{}{"website_group": webs},
		)
	}
}

func createWebsiteHandler(r repository.Repostory, conf *config.Config) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// userUUID, err := UserUUID(req)
		userUUID := req.Context().Value("userUUID").(string)
		url := req.Context().Value("webURL").(string)

		web := model.NewWebsite(url, conf)
		service.Update(context.Background(), r, &web)

		err := r.CreateWebsite(&web)
		if err != nil {
			log.Printf("path: %s; user_uuid: %s; web_url: %s; error: %s", req.URL, userUUID, url, err)
			writeError(res, http.StatusBadRequest, err)
			return
		}

		userWeb := model.NewUserWebsite(web, userUUID)
		err = r.CreateUserWebsite(&userWeb)
		if err != nil {
			log.Printf("path: %s; user_uuid: %s; web_url: %s; error: %s", req.URL, userUUID, url, err)
			writeError(res, http.StatusBadRequest, err)
			return
		}
		log.Printf("path: %s; user_uuid: %s; web_url: %s", req.URL, userUUID, url)

		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": fmt.Sprintf("website <%v> inserted", web.Title),
		})
	}
}

func getWebsiteHandler(r repository.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(model.UserWebsite)
		log.Printf("path: %s; web_uuid: %s;", req.URL, web.WebsiteUUID)
		json.NewEncoder(res).Encode(map[string]interface{}{
			"website": web,
		})
	}
}

func refreshWebsiteHandler(r repository.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(model.UserWebsite)
		log.Printf("path: %s; web_uuid: %s;", req.URL, web.WebsiteUUID)
		web.AccessTime = time.Now()
		err := r.UpdateUserWebsite(&web)
		if err != nil {
			writeError(res, http.StatusInternalServerError, err)
			return
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"website": web,
		})
	}
}

func deleteWebsiteHandler(r repository.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(model.UserWebsite)
		err := r.DeleteUserWebsite(&web)
		if err != nil {
			log.Printf("path: %s; web_uuid: %s; error: %s", req.URL, web.WebsiteUUID, err)
			writeError(res, http.StatusInternalServerError, err)
			return
		}
		log.Printf("path: %s; web_uuid: %s;", req.URL, web.WebsiteUUID)
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": fmt.Sprintf("website <%v> deleted", web.Website.Title),
		})
	}
}

func validGroupName(web model.UserWebsite, groupName string) bool {
	for _, char := range strings.Split(groupName, "") {
		if strings.Contains(web.Website.Title, char) {
			return true
		}
	}
	return false
}

func changeWebsiteGroupHandler(r repository.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(model.UserWebsite)
		groupName := req.Context().Value("group").(string)
		if !validGroupName(web, groupName) {
			writeError(res, http.StatusBadRequest, errors.New("invalid group name"))
			return
		}
		web.GroupName = groupName
		err := r.UpdateUserWebsite(&web)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"website": web,
		})
	}
}

func dbStatsHandler(r repository.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		json.NewEncoder(res).Encode(r.Stats())
	}
}
