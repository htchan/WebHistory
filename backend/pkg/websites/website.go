package websites

import (
	"time"
	"net/http"
	"io/ioutil"
	"strings"
	"regexp"

	"github.com/htchan/WebHistory/internal/logging"

	"github.com/htchan/ApiParser"
)

func logDates(url string, dates []string) {
	if len(dates) > 10 {
		logging.LogUpdate(url, dates[:10])
	} else {
		logging.LogUpdate(url, dates)
	}
}

type Website struct {
	UserUUID string
	Url, Title, GroupName string
	content string
	UpdateTime, AccessTime time.Time
}

const SEP = "\n"

func (website Website) _checkTimeUpdate(timeStr string) bool {
	if timeStr == "" { return false }
	layout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(layout, timeStr)
	if err == nil && t.After(website.UpdateTime) {
		return true
	}
	return false
}

func (website *Website) _checkBodyUpdate(responseBody string) bool {
	bodyUpdate, titleUpdate := false, false
	responseApi := ApiParser.Parse(responseBody, "website.info")
	title := responseApi.Data["Title"]
	dates := make([]string, len(responseApi.Items))
	for i := range responseApi.Items { dates[i] = responseApi.Items[i]["Date"] }
	logDates(website.Url, dates)
	if website.isUpdated(dates) {
		website.content = strings.Join(dates, SEP)
		bodyUpdate = true
	}
	if (title != website.Title) {
		website.Title = title
		titleUpdate = true
	}
	logging.LogUpdate(website.Url, website.Title)
	if (website.GroupName == "") {
		website.GroupName = website.Title
	}
	return bodyUpdate || titleUpdate
}

func (website *Website) isUpdated(updatedDates []string) bool {
	currentDates := strings.Split(website.content, SEP)
	if len(currentDates) == 0 || len(updatedDates) == 0 || len(currentDates) != len(updatedDates) { return false }
	for i := range currentDates {
		if currentDates[i] != updatedDates[i] {
			return true
		}
	}
	return false
}

func pruneResponse(response *http.Response) string {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil { return "" }
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>|<path.*?/path>)")
	bodyStr := re.ReplaceAllString(
		strings.ReplaceAll(strings.ReplaceAll(string(body), "\r", ""), "\n", ""),
		"<delete/>",
	)
	re = regexp.MustCompile("<(/?title.*?)>")
	bodyStr = re.ReplaceAllString(bodyStr, "[$1]")
	re = regexp.MustCompile("(<.*?>)+")
	bodyStr = re.ReplaceAllString(bodyStr, SEP)
	re = regexp.MustCompile("\\[(/?title.*?)\\]")
	bodyStr = re.ReplaceAllString(bodyStr, "<$1>")
	return bodyStr
}

func (website *Website) Update() {
	client := http.Client{Timeout: 30*time.Second}
	resp, err := client.Get(website.Url);
	if err != nil { 
		website.Title = "Unknown"
		if (website.GroupName == "") { website.GroupName = website.Title; }
		return
	}
	body := pruneResponse(resp)
	if website._checkTimeUpdate(resp.Header.Get("last-modified")) ||
		website._checkBodyUpdate(body) {
		logging.LogUpdate(website.Url, website.Title+"\tupdate")
		if (website.GroupName == "") { website.GroupName = website.Title; }
		website.UpdateTime = time.Now()
	} else {
		logging.LogUpdate(website.Url, website.Title+"\tnot update")
	}
}

func (website Website) Map() map[string]interface{} {
	return map[string]interface{} {
		"url": website.Url,
		"title": website.Title,
		"groupName": website.GroupName,
		"updateTime": website.UpdateTime,
		"accessTime": website.AccessTime,
	}
}
