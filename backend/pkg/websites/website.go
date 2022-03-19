package websites

import (
	"time"
	"net/http"
	"io/ioutil"
	"strings"
	"regexp"

	"github.com/htchan/WebHistory/internal/logging"
)

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

func (website Website) getContent(client http.Client) string {
	resp, err := client.Get(website.Url)
	if err != nil { panic(err) }
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {panic(err)}
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>|<path.*?/path>)")
	bodyStr := string(re.ReplaceAll(
		[]byte(strings.ReplaceAll(strings.ReplaceAll(string(body), "\r", ""), "\n", "")),
		[]byte("<script/>")))
	return bodyStr
}

func (website Website) getTitle() string {
	client := http.Client{Timeout: 30*time.Second}
	body := website.getContent(client)
	re := regexp.MustCompile("<title.*?>(.*?)</title>")
	groups := re.FindStringSubmatch(body)
	if (len(groups) > 1) { return groups[1] }
	return ""
}

func (website *Website) _checkBodyUpdate(client http.Client, url string) bool {
	bodyUpdate, titleUpdate := false, false
	body := reduce(website.getContent(client), url)
	if !contentUpdated(website.content, body) {
		website.content = body
		bodyUpdate = true
	}
	title := website.getTitle()
	if (title != website.Title) {
		website.Title = title
		titleUpdate = true
	}
	print(website.GroupName)
	if (website.GroupName == "") {
		website.GroupName = website.Title
	}
	return bodyUpdate || titleUpdate
}

func contentUpdated(s1, s2 string) bool {
	list1, list2 := strings.Split(s1, SEP), strings.Split(s2, SEP)
	if len(list1) == 0 || len(list2) == 0 || len(list1) != len(list2) { return false }
	for i := range list1 {
		if list1[i] != list2[i] {
			return false
		}
	}
	return true
}

func extractContent(s string) string {
	re := regexp.MustCompile("<.*?>")
	return string(re.ReplaceAll([]byte(s), []byte(SEP)))
}

func extractDate(s, url string) string {
	re := regexp.MustCompile("\\d{1,4}([-/年月日號号]\\d{1,4}[年月日號号]?)+")
	resultList := re.FindAllString(s, -1)
	if len(resultList) > 10 {
		logging.LogUpdate(url, resultList[:10])
	} else {
		logging.LogUpdate(url, resultList)
	}
	return strings.Join(resultList, SEP)
}

func validDate(s string) string {
	validLength := 2
	dates := strings.Split(s, SEP)
	if validLength > len(dates) { validLength = len(dates) }
	return strings.Join(dates[:validLength], SEP)
}

func replaceKeyword(inputStr string, targetStr []string, replaceStr string) string {
	re := regexp.MustCompile("(" + strings.Join(targetStr, "|") + ")")
	return string(re.ReplaceAll([]byte(inputStr), []byte(replaceStr)))
}

func reduce(s, url string) (result string) {
	result = extractContent(s)
	result = validDate(extractDate(result, url))
	// print(result)
	return
}

func (website *Website) Update() {
	client := http.Client{Timeout: 30*time.Second}
	resp, err := client.Get(website.Url);
	if err != nil { 
		website.Title = "Unknown"
		if (website.GroupName == "") { website.GroupName = website.Title; }
		return
	}
	if website._checkTimeUpdate(resp.Header.Get("last-modified")) ||
		website._checkBodyUpdate(client, website.Url) {
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
