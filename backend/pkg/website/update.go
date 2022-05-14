package website

import (
	"time"
	"net/http"
	"strings"
	"log"
	"io/ioutil"
	"regexp"
	"github.com/htchan/ApiParser"
)


const SEP = "\n"
const DateLength = 10

func (web Website) _checkTimeUpdate(timeStr string) bool {
	if timeStr == "" { return false }
	layout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(layout, timeStr)
	if err == nil && t.After(web.UpdateTime) {
		return true
	}
	return false
}

func (web *Website) isUpdated(updatedDates []string) bool {
	currentDates := strings.Split(web.content, SEP)
	if len(updatedDates) == 0 {
		return false
	}
	if len(currentDates) < len(updatedDates) {
		return true
	}
	for i := range currentDates {
		if i >= DateLength {
			break
		}
		if currentDates[i] != updatedDates[i] {
			return true
		}
	}
	return false
}

func (web *Website) _checkBodyUpdate(responseBody string) bool {
	bodyUpdate, titleUpdate := false, false
	responseApi := ApiParser.Parse("website.info", responseBody)
	title := responseApi.Data["Title"]
	dates := make([]string, len(responseApi.Items))
	for i := range responseApi.Items {
		dates[i] = responseApi.Items[i]["Date"]
	}
	log.Println(web.URL, dates)
	if web.isUpdated(dates) {
		web.content = strings.Join(dates, SEP)
		bodyUpdate = true
	}
	if (title != web.Title) {
		web.Title = title
		titleUpdate = true
	}
	log.Println(web.URL, "title", web.Title)
	if (web.GroupName == "") {
		web.GroupName = web.Title
	}
	return bodyUpdate || titleUpdate
}

func pruneResponse(response *http.Response) string {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil { return "" }
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>|<path.*?/path>)")
	bodyStr := strings.ReplaceAll(strings.ReplaceAll(string(body), "\r", ""), "\n", "")
	bodyStr = re.ReplaceAllString(
		bodyStr,
		"<delete/>",
	)
	re = regexp.MustCompile("<(/?title.*?)>")
	bodyStr = re.ReplaceAllString(bodyStr, "[$1]")
	re = regexp.MustCompile("(<.*?>)+")
	bodyStr = re.ReplaceAllString(bodyStr, SEP)
	re = regexp.MustCompile("\\[(/?title.*?)\\]")
	bodyStr = re.ReplaceAllString(bodyStr, "<$1>")
	if strings.HasPrefix(bodyStr, SEP) {
		bodyStr = bodyStr[len(SEP):]
	}
	if strings.HasSuffix(bodyStr, SEP) {
		bodyStr = bodyStr[:len(bodyStr) - len(SEP)]
	}
	return bodyStr
}

func (web *Website) Update() {
	client := http.Client{Timeout: 30*time.Second}
	resp, err := client.Get(web.URL);
	if err != nil { 
		if web.Title == "" { web.Title = "Unknown" }
		if (web.GroupName == "") { web.GroupName = web.Title; }
		log.Println(web.URL, "failed to fetch", err)
		return
	}
	body := pruneResponse(resp)
	// if website._checkTimeUpdate(resp.Header.Get("last-modified")) ||
	if web._checkBodyUpdate(body) {
		log.Println(web.URL, "updated", web.Title)
		if (web.GroupName == "") { web.GroupName = web.Title; }
		web.UpdateTime = time.Now()
	} else {
		log.Println(web.URL, "not-updated", web.Title)
	}
}