package model

import (
	"encoding/json"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	KEY_SEPARATOR       = "WEB_WATCHER_SEPARATOR"
	KEY_DATE_MEX_LENGTH = "WEB_WATCHER_DATE_MAX_LENGTH"
)

var SEP = os.Getenv(KEY_SEPARATOR)
var DateLength, _ = strconv.Atoi(os.Getenv(KEY_DATE_MEX_LENGTH))

type Website struct {
	UUID       string
	URL        string
	Title      string
	RawContent string
	UpdateTime time.Time
}

func NewWebsite(url string) Website {
	web := Website{
		UUID:       uuid.New().String(),
		URL:        url,
		UpdateTime: time.Now(),
	}
	return web
}

func (web Website) Map() map[string]interface{} {
	return map[string]interface{}{
		"uuid":       web.UUID,
		"url":        web.URL,
		"title":      web.Title,
		"updateTime": web.UpdateTime,
	}
}

func (web Website) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UUID       string `json:"uuid"`
		URL        string `json:"url"`
		Title      string `json:"title"`
		UpdateTime string `json:"update_time"`
	}{
		UUID:       web.UUID,
		URL:        web.URL,
		Title:      web.Title,
		UpdateTime: web.UpdateTime.Format("2006-01-02T15:04:05 MST"),
	})
}

func (web Website) Host() string {
	u, err := url.Parse(web.URL)
	if err != nil || web.URL == "" {
		return ""
	}
	host := u.Host
	splitedHost := strings.Split(host, ".")
	return strings.Join(splitedHost[len(splitedHost)-2:], ".")
}

func (web Website) Content() []string {
	return strings.Split(web.RawContent, SEP)
}

func (web Website) Equal(compare Website) bool {
	return web.UUID == compare.UUID &&
		web.URL == compare.URL &&
		web.Title == compare.Title &&
		web.RawContent == compare.RawContent &&
		web.UpdateTime.Unix()/1000 == compare.UpdateTime.Unix()/1000
}
