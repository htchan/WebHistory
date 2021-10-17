package websites

import (
	"time"
	"net/http"
	"io/ioutil"
	"strings"
	"regexp"
	"fmt"
	
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

type Website struct {
	Url, Title, GroupName string
	content string
	UpdateTime, AccessTime time.Time
}

const SEP = "\n"

func OpenDatabase(location string) {
	var err error
	database, err = sql.Open("sqlite3", location)
	if err != nil { panic(err) }
	database.SetMaxIdleConns(5);
	database.SetMaxOpenConns(50);
	fmt.Println(database)
}

func closeDatabase() {
	database.Close()
}

func Urls() []string {
	resultUpdate := make([]string, 0)
	resultUnchange := make([]string, 0)
	rows, err := database.Query("select url, updateTime, accessTime from websites order by groupName, updateTime desc")
	if err != nil { panic(err) }
	var temp string
	var updateTime, accessTime int64
	for rows.Next() {
		rows.Scan(&temp, &updateTime, &accessTime)
		if updateTime > accessTime {
			resultUpdate = append(resultUpdate, temp)
		} else {
			resultUnchange = append(resultUnchange, temp)
		}
	}
	return append(resultUpdate, resultUnchange...)
}

func GroupNames() []string {
	result := make([]string, 0)
	rows, err := database.Query("select groupName from websites group by groupName order by max(updateTime) desc")
	if err != nil { panic(err) }
	var temp string
	for rows.Next() {
		rows.Scan(&temp)
		result = append(result, temp)
	}
	return result
}

func Group2Urls(groupName string) []string {
	urls := make([]string, 0)
	rows, err := database.Query("select url from websites where groupName=? order by updateTime desc", groupName)
	if err != nil { panic(err) }
	var temp string
	for rows.Next() {
		rows.Scan(&temp)
		urls = append(urls, temp)
	}
	return urls
}

func Url2Website(url string) Website {
	rows, err := database.Query(
		"select url, title, groupName, content, updateTime, accessTime from websites " +
		"where url=?", url)
	if err != nil { panic(err) }
	var web Website
	var updateTime, accessTime int
	if rows.Next() {
		rows.Scan(&web.Url, &web.Title, &web.GroupName, &web.content, &updateTime, &accessTime)
		web.UpdateTime = time.Unix(int64(updateTime), 0)
		web.AccessTime = time.Unix(int64(accessTime), 0)
	}
	err = rows.Close()
	if err != nil { panic(err) }
	return web
}

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
	body := reduce(website.getContent(client))
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

func extractDate(s string) string {
	re := regexp.MustCompile("\\d{1,4}([-/年月日號号]\\d{1,4}[年月日號号]?)+")
	resultList := re.FindAllString(s, -1)
	fmt.Println(resultList)
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

func reduce(s string) (result string) {
	result = extractContent(s)
	result = validDate(extractDate(result))
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
		fmt.Println(website.Title + "\tupdate")
		if (website.GroupName == "") { website.GroupName = website.Title; }
		website.UpdateTime = time.Now()
	} else {
		fmt.Println(website.Title + "\tnot update")
	}
}

func (website Website) insert(tx *sql.Tx) {
	_, err := tx.Exec("insert into websites (url, title, groupName, content, updateTime, accessTime) " +
		"values (?, ?, ?, ?, ?, ?)",
		website.Url, website.Title, website.GroupName, website.content, website.UpdateTime.Unix(), website.AccessTime.Unix())
	if err != nil { panic(err) }
}

func (website Website) Save() {
	tx, err := database.Begin()
	if err != nil { panic(err) }
	result, err := tx.Exec("update websites set " +
		"title=?, groupName=?, content=?, updateTime=?, accessTime=? where url=?",
		website.Title, website.GroupName, website.content, 
		website.UpdateTime.Unix(), website.AccessTime.Unix(), website.Url)
	if err != nil { panic(err) }
	rowsAffected, err := result.RowsAffected()
	if err != nil { panic(err) }
	if rowsAffected == 0 { website.insert(tx) }
	err = tx.Commit()
	if err != nil { panic(err) }
}

func (website Website) Delete() {
	tx, err := database.Begin()
	if err != nil { panic(err) }
	_, err = tx.Exec("delete from websites where url=?", website.Url)
	if err != nil { panic(err) }
	err = tx.Commit()
	if err != nil { panic(err) }

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
