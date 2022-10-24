package model

import (
	"github.com/htchan/ApiParser"
)

type WebsiteSetting struct {
	Domain         string
	TitleRegex     string
	ContentRegex   string
	FocusIndexFrom int
	FocusIndexTo   int
}

func (setting *WebsiteSetting) Parse(response string) (string, []string) {
	responseApi := ApiParser.Parse(setting.Domain, response)
	title := responseApi.Data["Title"]
	contents := make([]string, len(responseApi.Items))
	for i := range responseApi.Items {
		contents[i] = responseApi.Items[i]["Content"]
	}

	if setting.FocusIndexFrom < 0 {
		setting.FocusIndexFrom = (setting.FocusIndexFrom + len(contents)) % len(contents)
	}

	if setting.FocusIndexTo == 0 {
		setting.FocusIndexTo = len(contents)
	} else if setting.FocusIndexTo < 0 {
		setting.FocusIndexTo = (setting.FocusIndexTo + len(contents)) % len(contents)
	}
	if setting.FocusIndexFrom <= setting.FocusIndexTo {
		contents = contents[setting.FocusIndexFrom:setting.FocusIndexTo]
	}

	return title, contents
}
