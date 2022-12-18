package model

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/htchan/ApiParser"
)

type WebsiteSetting struct {
	Domain               string
	TitleRegex           string
	ContentRegex         string
	TitleGoquerySelector string
	DatesGoquerySelector string
	FocusIndexFrom       int
	FocusIndexTo         int
}

func (setting WebsiteSetting) Parse(response string) (string, []string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(response))
	if err != nil {
		fmt.Println("fail to parse error", err)

		return "", nil
	}

	title := doc.Find(setting.TitleGoquerySelector).Text()

	var dates []string
	doc.Find(setting.DatesGoquerySelector).Each(func(i int, s *goquery.Selection) {
		dates = append(dates, s.Text())
	})

	if len(dates) > 0 && setting.FocusIndexFrom < 0 {
		setting.FocusIndexFrom = (setting.FocusIndexFrom + len(dates)) % len(dates)
	} else if setting.FocusIndexFrom > len(dates) {
		setting.FocusIndexFrom = 0
	}

	if setting.FocusIndexTo == 0 {
		setting.FocusIndexTo = len(dates)
	} else if len(dates) > 0 && setting.FocusIndexTo < 0 {
		setting.FocusIndexTo = (setting.FocusIndexTo + len(dates)) % len(dates)
	} else if setting.FocusIndexTo > len(dates) {
		setting.FocusIndexTo = len(dates)
	}

	if len(dates) > 0 && setting.FocusIndexFrom <= setting.FocusIndexTo {
		dates = dates[setting.FocusIndexFrom:setting.FocusIndexTo]
	}

	return title, dates
}

func (setting *WebsiteSetting) ParseOld(response string) (string, []string) {
	responseApi := ApiParser.Parse(setting.Domain, response)
	title := responseApi.Data["Title"]
	contents := make([]string, len(responseApi.Items))
	for i := range responseApi.Items {
		contents[i] = responseApi.Items[i]["Content"]
	}

	if len(contents) > 0 && setting.FocusIndexFrom < 0 {
		setting.FocusIndexFrom = (setting.FocusIndexFrom + len(contents)) % len(contents)
	} else if setting.FocusIndexFrom > len(contents) {
		setting.FocusIndexFrom = 0
	}

	if setting.FocusIndexTo == 0 {
		setting.FocusIndexTo = len(contents)
	} else if len(contents) > 0 && setting.FocusIndexTo < 0 {
		setting.FocusIndexTo = (setting.FocusIndexTo + len(contents)) % len(contents)
	} else if setting.FocusIndexTo > len(contents) {
		setting.FocusIndexTo = len(contents)
	}

	if len(contents) > 0 && setting.FocusIndexFrom <= setting.FocusIndexTo {
		contents = contents[setting.FocusIndexFrom:setting.FocusIndexTo]
	}

	return title, contents
}
