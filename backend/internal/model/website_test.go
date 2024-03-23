package model

import (
	"strings"
	"testing"
	"time"

	"github.com/htchan/WebHistory/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_NewWebsite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		conf *config.WebsiteConfig
		want Website
	}{
		{
			name: "happy flow",
			url:  "https://google.com",
			conf: &config.WebsiteConfig{
				Separator:     "\n",
				MaxDateLength: 3,
			},
			want: Website{
				UUID:       "",
				URL:        "https://google.com",
				Title:      "",
				RawContent: "",
				UpdateTime: time.Now().UTC().Truncate(time.Second),
				Conf:       &config.WebsiteConfig{Separator: "\n", MaxDateLength: 3},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := NewWebsite(test.url, test.conf)

			// assert websites
			assert.NotEmpty(t, got.UUID)
			got.UUID = ""
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_NewWebsiteFromMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    map[string]interface{}
		want Website
	}{
		{
			name: "happy flow",
			m: map[string]interface{}{
				"uuid":            "uuid",
				"url":             "http://example.com",
				"title":           "title",
				"raw_content":     "content\ncontent\ncontent",
				"update_time":     time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local),
				"separator":       "\n",
				"max_date_length": 3,
			},
			want: Website{
				UUID:       "uuid",
				URL:        "http://example.com",
				Title:      "title",
				RawContent: "content\ncontent\ncontent",
				UpdateTime: time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local),
				Conf: &config.WebsiteConfig{
					Separator:     "\n",
					MaxDateLength: 3,
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := NewWebsiteFromMap(test.m)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestWebsite_Map(t *testing.T) {
	tests := []struct {
		name string
		web  Website
		want map[string]interface{}
	}{
		{
			name: "happy flow",
			web: Website{
				UUID:       "uuid",
				URL:        "http://example.com",
				Title:      "title",
				RawContent: "content\ncontent\ncontent",
				UpdateTime: time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local),
				Conf: &config.WebsiteConfig{
					Separator:     "\n",
					MaxDateLength: 3,
				},
			},
			want: map[string]interface{}{
				"uuid":            "uuid",
				"url":             "http://example.com",
				"title":           "title",
				"raw_content":     "content\ncontent\ncontent",
				"update_time":     time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local),
				"separator":       "\n",
				"max_date_length": 3,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.web.Map()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestWebsite_Host(t *testing.T) {
	tests := []struct {
		name string
		web  Website
		want string
	}{
		{
			name: "happy flow",
			web:  Website{URL: "http://example.com"},
			want: "example.com",
		},
		{
			name: "fail flow",
			web:  Website{URL: ""},
			want: "",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.web.Host()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestWebsite_Content(t *testing.T) {
	tests := []struct {
		name string
		web  Website
		want []string
	}{
		{
			name: "happy flow",
			web: Website{
				RawContent: strings.Join([]string{"1", "2", "3"}, "\n"),
				Conf:       &config.WebsiteConfig{Separator: "\n"},
			},
			want: []string{"1", "2", "3"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.web.Content()
			assert.Equal(t, test.want, got)
		})
	}
}
