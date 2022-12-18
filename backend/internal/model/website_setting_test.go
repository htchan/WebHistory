package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/go-cmp/cmp"
	"github.com/htchan/ApiParser"
)

func TestWebsiteSetting_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setting     *WebsiteSetting
		resp        string
		expectTitle string
		expectDates []string
	}{
		{
			name: "happy flow",
			setting: &WebsiteSetting{
				TitleGoquerySelector: "head>title",
				DatesGoquerySelector: "ul>li",
			},
			resp:        "<html><head><title>test</title></head><body><ul><li>1</li><li>2</li></ul></body></html>",
			expectTitle: "test",
			expectDates: []string{"1", "2"},
		},
		{
			name: "fail parse resp to doc",
			setting: &WebsiteSetting{
				TitleGoquerySelector: "head>title",
				DatesGoquerySelector: "ul>li",
			},
			resp:        "some unknown resp",
			expectTitle: "",
			expectDates: nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			title, dates := test.setting.Parse(test.resp)
			assert.Equal(t, test.expectTitle, title)
			assert.Equal(t, test.expectDates, dates)
		})
	}
}

func TestWebsiteSetting_ParseOld(t *testing.T) {
	tests := []struct {
		name          string
		setting       *WebsiteSetting
		resp          string
		expectTitle   string
		expectContent []string
	}{
		{
			name:          "works",
			setting:       &WebsiteSetting{Domain: "hello", TitleRegex: "(?P<Title>title\\d+)", ContentRegex: "(?P<Content>cont\\d+)"},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont1", "cont2", "cont3"},
		},
		{
			name: "works with specific range",
			setting: &WebsiteSetting{
				Domain:         "hello",
				TitleRegex:     "(?P<Title>title\\d+)",
				ContentRegex:   "(?P<Content>cont\\d+)",
				FocusIndexFrom: 1,
				FocusIndexTo:   2,
			},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont2"},
		},
		{
			name: "works with specific start only",
			setting: &WebsiteSetting{
				Domain:         "hello",
				TitleRegex:     "(?P<Title>title\\d+)",
				ContentRegex:   "(?P<Content>cont\\d+)",
				FocusIndexFrom: 2,
			},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont3"},
		},
		{
			name: "works with negative start",
			setting: &WebsiteSetting{
				Domain:         "hello",
				TitleRegex:     "(?P<Title>title\\d+)",
				ContentRegex:   "(?P<Content>cont\\d+)",
				FocusIndexFrom: -1,
			},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont3"},
		},
		{
			name: "works with specific end",
			setting: &WebsiteSetting{
				Domain:       "hello",
				TitleRegex:   "(?P<Title>title\\d+)",
				ContentRegex: "(?P<Content>cont\\d+)",
				FocusIndexTo: 2,
			},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont1", "cont2"},
		},
		{
			name: "works with zero end",
			setting: &WebsiteSetting{
				Domain:       "hello",
				TitleRegex:   "(?P<Title>title\\d+)",
				ContentRegex: "(?P<Content>cont\\d+)",
				FocusIndexTo: 0,
			},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont1", "cont2", "cont3"},
		},
		{
			name: "works with negative end",
			setting: &WebsiteSetting{
				Domain:       "hello",
				TitleRegex:   "(?P<Title>title\\d+)",
				ContentRegex: "(?P<Content>cont\\d+)",
				FocusIndexTo: -1,
			},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont1", "cont2"},
		},
		{
			name: "return all content with invalid range",
			setting: &WebsiteSetting{
				Domain:         "hello",
				TitleRegex:     "(?P<Title>title\\d+)",
				ContentRegex:   "(?P<Content>cont\\d+)",
				FocusIndexFrom: 2,
				FocusIndexTo:   -2,
			},
			resp:          "title123 cont1 cont2 cont3",
			expectTitle:   "title123",
			expectContent: []string{"cont1", "cont2", "cont3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ApiParser.SetDefault(ApiParser.NewFormatSet(test.setting.Domain, test.setting.ContentRegex, test.setting.TitleRegex))
			title, content := test.setting.ParseOld(test.resp)

			if title != test.expectTitle {
				t.Errorf("got title: %v; want title: %v", title, test.expectTitle)
			}
			if !cmp.Equal(content, test.expectContent) {
				t.Errorf("content diff: %v", cmp.Diff(content, test.expectContent))
			}
		})
	}
}
