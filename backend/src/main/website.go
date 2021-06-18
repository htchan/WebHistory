package main

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
	Url, Title string
	content string
	UpdateTime, AccessTime time.Time
}

const SEP = "\n"

func openDatabase(location string) {
	var err error
	database, err = sql.Open("sqlite3", location)
	if err != nil { panic(err) }
	database.SetMaxIdleConns(5);
	database.SetMaxOpenConns(50);
}

func closeDatabase() {
	database.Close()
}

func Urls() []string {
	result := make([]string, 0)
	rows, err := database.Query("select url from websites order by updateTime desc")
	if err != nil { panic(err) }
	var temp string
	for rows.Next() {
		rows.Scan(&temp)
		result = append(result, temp)
	}
	return result
}

func GetWeb(url string) Website {
	rows, err := database.Query(
		"select url, title, content, updateTime, accessTime from websites " +
		"where url=?", url)
	if err != nil { panic(err) }
	var web Website
	var updateTime, accessTime int
	if rows.Next() {
		rows.Scan(&web.Url, &web.Title, &web.content, &updateTime, &accessTime)
		web.UpdateTime = time.Unix(int64(updateTime), 0)
		web.AccessTime = time.Unix(int64(accessTime), 0)
	}
	err = rows.Close()
	if err != nil { panic(err) }
	return web
}

func (web *Website) _checkTimeUpdate(timeStr string) bool {
	if timeStr == "" { return false }
	layout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(layout, timeStr)
	if err == nil && t.After(web.UpdateTime) {
		return true
	}
	return false
}

func getContent(client http.Client, url string) string {
	resp, err := client.Get(url)
	if err != nil { panic(err) }
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {panic(err)}
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>|<path.*?/path>)")
	bodyStr := string(re.ReplaceAll(
		[]byte(strings.ReplaceAll(strings.ReplaceAll(string(body), "\r", ""), "\n", "")),
		[]byte("<script/>")))
	return bodyStr
}

func getTitle(url string) string {
	client := http.Client{Timeout: 30*time.Second}
	body := getContent(client, url)
	re := regexp.MustCompile("<title.*?>(.*?)</title>")
	groups := re.FindStringSubmatch(body)
	if (len(groups) > 1) { return groups[1] }
	return ""
}

func (web *Website) _checkBodyUpdate(client http.Client, url string) bool {
	bodyUpdate, titleUpdate := false, false
	body := reduce(getContent(client, url))
	if !compare(web.content, body) {
		web.content = body
		bodyUpdate = true
	}
	title := getTitle(url)
	if (title != web.Title) {
		web.Title = title
		titleUpdate = true
	}
	return bodyUpdate || titleUpdate
}

func compare(s1, s2 string) bool {
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
	result := string(re.ReplaceAll([]byte(s), []byte(SEP)))
	fmt.Println(resultList)
	return strings.Join(resultList, SEP)
}

func extractDate(s string) string {
	re := regexp.MustCompile("<.*?>")
	result := string(re.ReplaceAll([]byte(s), []byte(SEP)))
	re = regexp.MustCompile("\\d+[-/][-/\\d]+")
	resultList := re.FindAllString(result, -1)
	fmt.Println(resultList)
	return strings.Join(resultList, SEP)
}

func replaceKeyword(s []string, replaceStr string) string {
	re := regexp.MustCompile("(" + strings.Join(s, "|") + ")")
	result := string(re.ReplaceAll([]byte(s), []byte(replaceStr)))
	fmt.Println(resultList)
	return strings.Join(resultList, SEP)
}

func reduce(s string) (result string) {
	result = strings.Join(strings.Split(replaceDate, SEP)[:2], SEP)
	return
}

func (web *Website) Update() {
	client := http.Client{Timeout: 30*time.Second}
	resp, err := client.Get(web.Url);
	if err != nil { 
		web.Title = "Unknown"
		return
	}
	if web._checkTimeUpdate(resp.Header.Get("last-modified")) ||
		web._checkBodyUpdate(client, web.Url) {
		fmt.Println(web.Title + "\tupdate")
		web.UpdateTime = time.Now()
	} else {
		fmt.Println(web.Title + "\tnot update")
	}
}

func (web Website) insert(tx *sql.Tx) {
	_, err := tx.Exec("insert into websites (url, title, content, updateTime, accessTime) values (?, ?, ?, ?, ?)",
		web.Url, web.Title, web.content, web.UpdateTime.Unix(), web.AccessTime.Unix())
	if err != nil { panic(err) }
}

func (web Website) Save() {
	tx, err := database.Begin()
	if err != nil { panic(err) }
	result, err := tx.Exec("update websites set title=?, content=?, updateTime=?, accessTime=? where url=?",
		web.Title, web.content, web.UpdateTime.Unix(), web.AccessTime.Unix(), web.Url)
	if err != nil { panic(err) }
	rowsAffected, err := result.RowsAffected()
	if err != nil { panic(err) }
	if rowsAffected == 0 { web.insert(tx) }
	err = tx.Commit()
	if err != nil { panic(err) }
}

func (web Website) Delete() {
	tx, err := database.Begin()
	if err != nil { panic(err) }
	_, err = tx.Exec("delete from websites where url=?", web.Url)
	if err != nil { panic(err) }
	err = tx.Commit()
	if err != nil { panic(err) }

}

func (web Website) Response() map[string]interface{} {
	return map[string]interface{} {
		"url": web.Url,
		"title": web.Title,
		"updateTime": web.UpdateTime,
		"accessTime": web.AccessTime,
	}
}