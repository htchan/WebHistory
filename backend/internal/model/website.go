package model

import (
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/htchan/WebHistory/internal/config"
	"go.opentelemetry.io/otel/attribute"
)

type Website struct {
	UUID       string    `json:"uuid"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	RawContent string    `json:"raw_content"`
	UpdateTime time.Time `json:"update_time"`
	Conf       *config.WebsiteConfig
}

func NewWebsiteFromMap(m map[string]interface{}) Website {
	return Website{
		UUID:       m["uuid"].(string),
		URL:        m["url"].(string),
		Title:      m["title"].(string),
		RawContent: m["raw_content"].(string),
		UpdateTime: m["update_time"].(time.Time),
		Conf: &config.WebsiteConfig{
			Separator:     m["separator"].(string),
			MaxDateLength: m["max_date_length"].(int),
		},
	}
}

func NewWebsite(url string, conf *config.WebsiteConfig) Website {
	web := Website{
		UUID:       uuid.New().String(),
		URL:        url,
		UpdateTime: time.Now().UTC().Truncate(time.Second),
		Conf:       conf,
	}
	return web
}

func (web Website) Map() map[string]interface{} {
	return map[string]interface{}{
		"uuid":            web.UUID,
		"url":             web.URL,
		"title":           web.Title,
		"raw_content":     web.RawContent,
		"update_time":     web.UpdateTime,
		"separator":       web.Conf.Separator,
		"max_date_length": web.Conf.MaxDateLength,
	}
}

func (web Website) Host() string {
	u, err := url.Parse(web.URL)
	if err != nil || web.URL == "" {
		return ""
	}
	splitedHost := strings.Split(u.Hostname(), ".")
	return strings.Join(splitedHost[len(splitedHost)-2:], ".")
}

func (web Website) Content() []string {
	return strings.Split(web.RawContent, web.Conf.Separator)
}

func (web Website) Equal(compare Website) bool {
	return web.UUID == compare.UUID &&
		web.URL == compare.URL &&
		web.Title == compare.Title &&
		web.RawContent == compare.RawContent &&
		web.UpdateTime.Unix()/1000 == compare.UpdateTime.Unix()/1000
}

func (web Website) OtelAttributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("url", web.URL),
		attribute.String("title", web.Title),
	}
}
