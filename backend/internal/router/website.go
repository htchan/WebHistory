package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repo"
	"github.com/htchan/WebHistory/internal/service"
	"github.com/htchan/WebHistory/internal/utils"
)

func writeError(res http.ResponseWriter, statusCode int, err error) {
	res.WriteHeader(statusCode)
	fmt.Fprintln(res, fmt.Sprintf(`{ "error": "%v" }`, err))
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
		func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(res, req)
		},
	)
}

func getAllWebsiteGroups(r repo.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		webs, err := r.FindUserWebsites(userUUID)
		if err != nil {
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
			return
		}

		json.NewEncoder(res).Encode(map[string]interface{}{
			"website_groups": webs.WebsiteGroups(),
		})
	}
}

func getWebsiteGroup(r repo.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userUUID := req.Context().Value("userUUID").(string)
		groupName := chi.URLParam(req, "groupName")
		webs, err := r.FindUserWebsitesByGroup(userUUID, groupName)
		if err != nil || len(webs) == 0 {
			writeError(res, http.StatusBadRequest, RecordNotFoundError)
			return
		}

		json.NewEncoder(res).Encode(
			map[string]interface{}{"website_group": webs},
		)
	}
}

func WebsiteParams(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
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

func createWebsite(r repo.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// userUUID, err := UserUUID(req)
		userUUID := req.Context().Value("userUUID").(string)
		url := req.Context().Value("webURL").(string)

		web := model.NewWebsite(url)
		service.Update(r, &web)

		err := r.CreateWebsite(&web)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}

		userWeb := model.NewUserWebsite(web, userUUID)
		err = r.CreateUserWebsite(&userWeb)
		if err != nil {
			writeError(res, http.StatusBadRequest, err)
			return
		}

		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": fmt.Sprintf("website <%v> inserted", web.Title),
		})
	}
}

func QueryWebsite(r repo.Repostory) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				userUUID := req.Context().Value("userUUID").(string)
				webUUID := chi.URLParam(req, "webUUID")
				web, err := r.FindUserWebsite(userUUID, webUUID)
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

func getWebsite(r repo.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(model.UserWebsite)
		json.NewEncoder(res).Encode(map[string]interface{}{
			"website": web,
		})
	}
}

func refreshWebsite(r repo.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(model.UserWebsite)
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

func deleteWebsite(r repo.Repostory) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		web := req.Context().Value("website").(model.UserWebsite)
		err := r.DeleteUserWebsite(&web)
		if err != nil {
			writeError(res, http.StatusInternalServerError, err)
			log.Println("delete-website", err)
			return
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": fmt.Sprintf("website <%v> deleted", web.Website.Title),
		})
	}
}

func GroupNameParams(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
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

func validGroupName(web model.UserWebsite, groupName string) bool {
	for _, char := range strings.Split(groupName, "") {
		if strings.Contains(web.Website.Title, char) {
			return true
		}
	}
	return false
}

func changeWebsiteGroup(r repo.Repostory) http.HandlerFunc {
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

func AddWebsiteRoutes(router chi.Router, r repo.Repostory) {
	api_route_prefix := os.Getenv("WEB_WATCHER_API_ROUTE_PREFIX")
	if api_route_prefix == "" {
		api_route_prefix = "/api/web-watcher"
	}
	router.Route(api_route_prefix, func(router chi.Router) {
		router.Route("/websites", func(router chi.Router) {
			router.Use(
				cors.Handler(
					cors.Options{
						AllowedOrigins: []string{"*"},
						AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
						AllowedHeaders: []string{"*"},
						MaxAge:         300, // Maximum value not ignored by any of major browsers
					},
				),
			)
			router.Use(Authenticate)
			router.Use(SetContentType)

			router.Route("/groups", func(router chi.Router) {
				router.Get("/", getAllWebsiteGroups(r))
				router.Get("/{groupName}", getWebsiteGroup(r))
			})

			router.With(WebsiteParams).Post("/", createWebsite(r))

			router.With(QueryWebsite(r)).Route("/{webUUID}", func(router chi.Router) {
				router.Get("/", getWebsite(r))
				router.Delete("/", deleteWebsite(r))
				router.Put("/refresh", refreshWebsite(r))
				router.With(GroupNameParams).Put("/change-group", changeWebsiteGroup(r))
			})
		})
	})
}
