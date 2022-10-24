package model

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/htchan/ApiParser"
)

func TestWebsiteSetting_parse(t *testing.T) {
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
			title, content := test.setting.Parse(test.resp)

			if title != test.expectTitle {
				t.Errorf("got title: %v; want title: %v", title, test.expectTitle)
			}
			if !cmp.Equal(content, test.expectContent) {
				t.Errorf("content diff: %v", cmp.Diff(content, test.expectContent))
			}
		})
	}
}
