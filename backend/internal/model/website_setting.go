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

func (setting *WebsiteSetting) Parse(response string) (string, []string) {
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

	fromN, toN := setting.FocusIndexFrom, setting.FocusIndexTo
	if fromN < 0 {
		fromN = len(dates) + fromN
		if fromN < 0 {
			fromN = 0
		}
	} else if fromN > len(dates) {
		fromN = len(dates) - 1
	}

	if toN <= 0 {
		toN = len(dates) + toN
		if toN < 0 {
			toN = len(dates)
		}
	} else if toN > len(dates) {
		toN = len(dates)
	}
	if fromN <= toN {
		dates = dates[fromN:toN]
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

	fromN, toN := setting.FocusIndexFrom, setting.FocusIndexTo
	if fromN < 0 {
		fromN = len(contents) + fromN
		if fromN < 0 {
			fromN = 0
		}
	} else if fromN > len(contents) {
		fromN = len(contents) - 1
	}

	if toN <= 0 {
		toN = len(contents) + toN
		if toN < 0 {
			toN = len(contents)
		}
	} else if toN > len(contents) {
		toN = len(contents)
	}
	if fromN <= toN {
		contents = contents[fromN:toN]
	}

	return title, contents
}
