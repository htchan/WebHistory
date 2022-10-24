package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repo"
)

type HTTPClient interface {
	Get(string) (*http.Response, error)
}

var client HTTPClient = &http.Client{Timeout: 30 * time.Second}

func pruneResponse(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile("[\r|\n|\t]")
	bodyStr := re.ReplaceAllString(string(body), "")
	re = regexp.MustCompile("(<script.*?/script>|<style.*?/style>|<path.*?/path>)")
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

func getWebsiteSetting(r repo.Repostory, web *model.Website) (*model.WebsiteSetting, error) {
	u, err := url.Parse(web.URL)
	setting, err := r.FindWebsiteSetting(u.Hostname())
	if err == nil {
		return setting, nil
	}
	return r.FindWebsiteSetting("default")
}

func parseAPI(r repo.Repostory, web *model.Website, resp string) (string, []string) {
	setting, err := getWebsiteSetting(r, web)
	if err != nil {
		return "", nil
	}
	return setting.Parse(resp)
}

func fetchWebsite(web *model.Website) (string, error) {
	resp, err := client.Get(web.URL)
	if err != nil {
		if web.Title == "" {
			web.Title = "unknown"
		}
		log.Printf("url: %s; fail to fetch website; error: %s", web.URL, err)
		return "", fmt.Errorf("fail to fetch website response: %s", web.URL)
	}
	return pruneResponse(resp), nil
}
func checkTimeUpdated(web *model.Website, timeStr string) bool {
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

func checkContentUpdated(web *model.Website, content []string) bool {
	if len(content) > 0 && !cmp.Equal(web.Content(), content) {
		web.RawContent = strings.Join(content, model.SEP)
		web.UpdateTime = time.Now()
		// log.Printf("url: %s; content updated; new content: %s", web.URL, web.RawContent)
		return true
	}
	return false
}

func checkTitleUpdated(web *model.Website, title string) bool {
	if web.Title != title {
		if web.Title == "" || web.Title == "unknown" {
			web.Title = title
			web.UpdateTime = time.Now()
		}
		// log.Printf("url: %s; title updated; new title: %s", web.URL, web.Title)
		return true
	}
	return false
}

func Update(r repo.Repostory, web *model.Website) error {
	// log.Printf("url: %s, start", web.URL)
	content, err := fetchWebsite(web)
	if err != nil {
		return err
	}

	title, dates := parseAPI(r, web, content)
	titleUpdated := checkTitleUpdated(web, title)
	contentUpadted := checkContentUpdated(web, dates)

	if titleUpdated || contentUpadted {
		err := r.UpdateWebsite(web)
		if err != nil {
			// log.Printf("url: %s; website update failed; error: %s", web.URL, err)
		} else {
			// log.Printf("url: %s; website updated", web.URL)
		}
	}

	log.Printf("url: %s, finish", web.URL)
	return nil
}
