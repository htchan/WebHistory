package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/htchan/ApiParser"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repo"
)

type HTTPClient interface {
	Get(string) (*http.Response, error)
}

var client HTTPClient = &http.Client{Timeout: 30 * time.Second}

func isTimeUpdated(web *model.Website, timeStr string) bool {
	if timeStr == "" {
		return false
	}
	layout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(layout, timeStr)
	if err == nil && t.After(web.UpdateTime) {
		return true
	}
	return false
}

func isContentUpdated(web *model.Website, dates []string) bool {
	if len(dates) == 0 {
		return false
	}

	for i := range web.Content() {
		if i >= model.DateLength {
			break
		}
		if web.Content()[i] != dates[i] {
			return true
		}
	}
	return false
}

func isTitleUpdated(web *model.Website, title string) bool {
	return web.Title != title
}

func pruneResponse(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>|<path.*?/path>)")
	bodyStr := strings.ReplaceAll(strings.ReplaceAll(string(body), "\r", ""), "\n", "")
	bodyStr = re.ReplaceAllString(
		bodyStr,
		"<delete/>",
	)
	re = regexp.MustCompile("<(/?title.*?)>")
	bodyStr = re.ReplaceAllString(bodyStr, "[$1]")
	re = regexp.MustCompile("(<.*?>)+")
	bodyStr = re.ReplaceAllString(bodyStr, model.SEP)
	re = regexp.MustCompile("\\[(/?title.*?)\\]")
	bodyStr = re.ReplaceAllString(bodyStr, "<$1>")
	if strings.HasPrefix(bodyStr, model.SEP) {
		bodyStr = bodyStr[len(model.SEP):]
	}
	if strings.HasSuffix(bodyStr, model.SEP) {
		bodyStr = bodyStr[:len(bodyStr)-len(model.SEP)]
	}
	return bodyStr
}

func Update(r repo.Repostory, web *model.Website) error {
	log.Printf("url: %s, start", web.URL)
	resp, err := client.Get(web.URL)
	if err != nil {
		if web.Title == "" {
			web.Title = "unknown"
		}
		log.Printf("url: %s; fail to fetch website; error: %s", web.URL, err)
		return fmt.Errorf("fail to fetch website response: %s", web.URL)
	}
	content := pruneResponse(resp)

	var (
		// dateUpdated    = false
		titleUpdated   = false
		contentUpdated = false
		responseApi    = ApiParser.Parse("website.info", content)
		title          = responseApi.Data["Title"]
		dates          = make([]string, len(responseApi.Items))
	)
	for i := range responseApi.Items {
		dates[i] = responseApi.Items[i]["Date"]
	}
	if len(dates) > model.DateLength {
		dates = dates[:model.DateLength]
	}

	if isContentUpdated(web, dates) {
		contentUpdated = true
		web.RawContent = strings.Join(dates, model.SEP)
		web.UpdateTime = time.Now()
		log.Printf("url: %s; content updated; new content: %s", web.URL, web.RawContent)
	}

	if isTitleUpdated(web, title) {
		titleUpdated = true
		if web.Title == "" || web.Title == "unknown" {
			web.Title = title
			web.UpdateTime = time.Now()
		}
		log.Printf("url: %s; title updated; new title: %s", web.URL, web.Title)
	}

	if titleUpdated || contentUpdated {
		err := r.UpdateWebsite(web)
		if err != nil {
			log.Printf("url: %s; website update failed; error: %s", web.URL, err)
		} else {
			log.Printf("url: %s; website updated", web.URL)
		}
	}

	log.Printf("url: %s, finish", web.URL)
	return nil
}
