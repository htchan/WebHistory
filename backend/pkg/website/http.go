package website

import (
	"errors"
	"net/http"
	"database/sql"
	"log"
	"fmt"
	"os"
	"strings"
	"encoding/json"
	"time"
	"context"
	"github.com/go-chi/chi/v5"
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

func redirectLogin(res http.ResponseWriter, req *http.Request) {
	loginURL := os.Getenv("LOGIN_URL")
	serviceUUID := os.Getenv("SERVICE_UUID")
	http.Redirect(res, req, fmt.Sprintf("%v?service=%v", loginURL, serviceUUID), 302)
}

var UnauthorizedError = errors.New("unauthorized")
var InvalidParamsError = errors.New("invalid params")
var RecordNotFoundError = errors.New("record not found")

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			token := req.Header.Get("Authorization")
			userUUID := ""
			if token != "" {
				userUUID = utils.FindUserByToken(token)
			}
			if userUUID == "" {
				writeError(res, http.StatusUnauthorized, UnauthorizedError)
				return
			}
			ctx := context.WithValue(req.Context(), "userUUID", userUUID)
			next.ServeHTTP(res, req.WithContext(ctx))
		},
	)
}
func SetContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func (res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(res, req)
		},
	)
}

// func listWebsiteHandler(db *sql.DB) http.HandlerFunc {
// 	return func (res http.ResponseWriter, req *http.Request) {
// 		res.Header().Set("Content-Type", "application/json; charset=utf-8")
// 		res.Header().Set("Access-Control-Allow-Origin", "*")
// 		// userUUID, err := UserUUID(req)
// 		if err != nil {
// 			redirectLogin(res, req)
// 			return
// 		}
// 		websites, err := FindAllUserWebsites(db, userUUID)
// 		if err != nil {
// 			writeError(res, http.StatusBadRequest, err)
// 			return
// 		}
// 		websiteGroups := WebsitesToWebsiteGroups(websites)
// 		err = json.NewEncoder(res).Encode(
// 			map[string][]WebsiteGroup{"website_groups": websiteGroups},
// 		)
// 		if err != nil {
// 			writeError(res, http.StatusInternalServerError, errors.New("parse response error"))
// 		}
// 	}
// }

func getAllWebsiteGroups(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		webs, err := FindAllUserWebsites(db, userUUID)
		if err != nil {
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
		}
		json.NewEncoder(res).Encode(WebsitesToWebsiteGroups(webs))
	}
}

func getWebsiteGroup(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		groupName := chi.URLParam(req, "groupName")

		webs, err := FindAllUserWebsites(db, userUUID)
		if err != nil {
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
		}
		for _, g := range WebsitesToWebsiteGroups(webs) {
			if g[0].GroupName == groupName {
				json.NewEncoder(res).Encode(g)
				return
			}
		}
		writeError(res, http.StatusBadRequest, RecordNotFoundError)
	}
}

func WebsiteParams(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func (res http.ResponseWriter, req *http.Request) {
			err := req.ParseForm()
			if err != nil {
				writeError(res, http.StatusBadRequest, InvalidParamsError)
				return
			}
			url := req.Form.Get("url")
			if url == "" || !strings.HasPrefix(url, "http") {
				writeError(res, http.StatusBadRequest, InvalidParamsError)
				return
			}
			ctx := context.WithValue(req.Context(), "webURL", url)
			next.ServeHTTP(res, req.WithContext(ctx))
		},
	)
}

func createWebsite(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		// userUUID, err := UserUUID(req)
		userUUID := req.Context().Value("userUUID").(string)
		url := req.Context().Value("webURL").(string)
		log.Println("create-website.start", fmt.Sprintf(`{ "url": "%v" }`, url))
		web := NewWebsite(url, userUUID)
		err := web.Create(db)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		log.Println("create-website.complete", web.Map())
		fmt.Fprintln(res, fmt.Sprintf(`{ "message": "website <%v> inserted" }`, web.Title))
	}
}

func QueryWebsite(db *sql.DB) func(http.Handler) http.Handler {
	return func (next http.Handler) http.Handler {
		return http.HandlerFunc(
			func (res http.ResponseWriter, req *http.Request) {
				userUUID := req.Context().Value("userUUID").(string)
				webUUID := chi.URLParam(req, "webUUID")
				web, err := FindUserWebsite(db, userUUID, webUUID)
				if err != nil {
					writeError(res, http.StatusBadRequest, err)
					return
				}
				ctx := context.WithValue(req.Context(), "website", web)
				next.ServeHTTP(res, req.WithContext(ctx))
			},
		)
	}
}

func getWebsite(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("websites").(Website)
		json.NewEncoder(res).Encode(web)
	}
}

func refreshWebsite(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(Website)
		web.AccessTime = time.Now()
		err := web.Save(db)
		if err != nil {
			writeError(res, http.StatusInternalServerError, err)
			return
		}
		json.NewEncoder(res).Encode(web)
	}
}

func deleteWebsite(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(Website)
		err := web.Delete(db)
		if err != nil {
			writeError(res, http.StatusInternalServerError, err)
			log.Println("delete-website", err)
			return
		}
	}
}

func GroupNameParams(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func (res http.ResponseWriter, req *http.Request) {
			err := req.ParseForm()
			if err != nil {
				writeError(res, http.StatusBadRequest, InvalidParamsError)
				return
			}
			groupName := req.Form.Get("group_name")
			ctx := context.WithValue(req.Context(), "group", groupName)
			next.ServeHTTP(res, req.WithContext(ctx))
		},
	)
}

func changeWebsiteGroup(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(Website)
		groupName := req.Context().Value("group").(string)
		if !utils.IsSubSet(web.Title, strings.ReplaceAll(groupName, " ", "")) || len(web.Title) == 0 {
			writeError(res, http.StatusBadRequest, errors.New("invalid group name"))
			return
		}
		web.GroupName = groupName
		err := web.Save(db)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		json.NewEncoder(res).Encode(web)
	}
}

func AddWebsiteRoutes(router chi.Router, db *sql.DB) {
	router.Route("/api/web-history", func (router chi.Router) {
		router.Options("/*", optionsHandler)

		router.Route("/websites", func (router chi.Router) {
			router.Use(Authenticate)
			router.Use(SetContentType)

			router.Route("/groups", func (router chi.Router) {
				router.Get("/", getAllWebsiteGroups(db))
				router.Get("/{groupName}", getWebsiteGroup(db))
			})

			router.With(WebsiteParams).Post("/", createWebsite(db))

			router.Route("/{webUUID}", func (router chi.Router) {
				router.Use(QueryWebsite(db))
				
				router.Get("/", getWebsite(db))
				router.Delete("/", deleteWebsite(db))
				router.Put("refresh", refreshWebsite(db))
				router.With(GroupNameParams).Put("change-group", changeWebsiteGroup(db))
			})
		})
	})
}