package router

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
	"github.com/go-chi/cors"
	"github.com/htchan/WebHistory/internal/utils"
	"github.com/htchan/WebHistory/pkg/website"
)

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
			if req.Method == http.MethodOptions {
				next.ServeHTTP(res, req)
				return
			}
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

func getAllWebsiteGroups(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		webs, err := website.FindAllUserWebsites(db, userUUID)
		if err != nil {
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
		}
		json.NewEncoder(res).Encode(map[string]interface{} {
			"website_groups": website.WebsitesToWebsiteGroups(webs),
		})
	}
}

func getWebsiteGroup(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		groupName := chi.URLParam(req, "groupName")

		webs, err := website.FindAllUserWebsites(db, userUUID)
		if err != nil {
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
		}
		for _, g := range website.WebsitesToWebsiteGroups(webs) {
			if g[0].GroupName == groupName {
				json.NewEncoder(res).Encode(map[string]interface{} {
					"website_group": g,
				})
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
		web := website.NewWebsite(url, userUUID)
		err := web.Create(db)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}
		log.Println("create-website.complete", web.Map())
		json.NewEncoder(res).Encode(map[string]interface{} {
			"message": fmt.Sprintf("website <%v> inserted", web.Title),
		})
	}
}

func QueryWebsite(db *sql.DB) func(http.Handler) http.Handler {
	return func (next http.Handler) http.Handler {
		return http.HandlerFunc(
			func (res http.ResponseWriter, req *http.Request) {
				userUUID := req.Context().Value("userUUID").(string)
				webUUID := chi.URLParam(req, "webUUID")
				web, err := website.FindUserWebsite(db, userUUID, webUUID)
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
		web := req.Context().Value("websites").(website.Website)
		json.NewEncoder(res).Encode(map[string]interface{} {
			"website": web,
		})
	}
}

func refreshWebsite(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(website.Website)
		web.AccessTime = time.Now()
		err := web.Save(db)
		if err != nil {
			writeError(res, http.StatusInternalServerError, err)
			return
		}
		json.NewEncoder(res).Encode(map[string]interface{} {
			"website": web,
		})
	}
}

func deleteWebsite(db *sql.DB) http.HandlerFunc {
	return func (res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(website.Website)
		err := web.Delete(db)
		if err != nil {
			writeError(res, http.StatusInternalServerError, err)
			log.Println("delete-website", err)
			return
		}
		json.NewEncoder(res).Encode(map[string]interface{} {
			"message": fmt.Sprintf("website <%v> deleted", web.Title),
		})
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
		web := req.Context().Value("website").(website.Website)
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
		json.NewEncoder(res).Encode(map[string]interface{} {
			"website": web,
		})
	}
}

func AddWebsiteRoutes(router chi.Router, db *sql.DB) {
	router.Route("/api/web-history", func (router chi.Router) {
		router.Route("/websites", func (router chi.Router) {
			router.Use(
				cors.Handler(
					cors.Options{
					AllowedOrigins:   []string{"*"},
					AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:   []string{"*"},
					MaxAge:           300, // Maximum value not ignored by any of major browsers
					},
				),
			)
			router.Use(Authenticate)
			router.Use(SetContentType)

			router.Route("/groups", func (router chi.Router) {
				router.Get("/", getAllWebsiteGroups(db))
				router.Get("/{groupName}", getWebsiteGroup(db))
			})

			router.With(WebsiteParams).Post("/", createWebsite(db))

			router.With(QueryWebsite(db)).Route("/{webUUID}", func (router chi.Router) {
				router.Get("/", getWebsite(db))
				router.Delete("/", deleteWebsite(db))
				router.Put("/refresh", refreshWebsite(db))
				router.With(GroupNameParams).Put("/change-group", changeWebsiteGroup(db))
			})
		})
	})
}