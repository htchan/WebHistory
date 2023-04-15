package website

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/repository"
	"github.com/htchan/WebHistory/internal/utils"
)

func AuthenticateMiddleware(conf *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodOptions {
					next.ServeHTTP(res, req)
					return
				}
				token := req.Header.Get("Authorization")
				userUUID := ""
				if token != "" {
					userUUID = utils.FindUserByToken(token, &conf.UserServiceConfig)
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
}
func SetContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(res, req)
		},
	)
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

func QueryWebsite(r repository.Repostory) func(http.Handler) http.Handler {
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
				ctx := context.WithValue(req.Context(), "website", *web)
				next.ServeHTTP(res, req.WithContext(ctx))
			},
		)
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
