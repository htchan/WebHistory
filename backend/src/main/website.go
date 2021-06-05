package main

import (
	"fmt"
	"time"
	"net/http"
	"io/ioutil"
	"strings"
	"regexp"
	
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

type Website struct {
	Url, Title string
	content string
	UpdateTime, AccessTime time.Time
}

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
	rows, err := database.Query("select url from websites")
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
		"where url=? order by updateTime desc", url)
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
	if err != nil { panic(err) }
	return string(body)
}

func getTitle(body string) string {
	re := regexp.MustCompile("<title>(.*?)</title>")
	groups := re.FindStringSubmatch(body)
	if (len(groups) > 1) { return groups[1] }
	return ""
}

func (web *Website) _checkBodyUpdate(client http.Client, url string) bool {
	bodyUpdate, titleUpdate := false, false
	bodys := make([]string, 4)
	bodys[0] = getContent(client, url)
	if !compare(web.content, bodys[0]) {
		for i := 1; i < 4; i++ {
			bodys[i] = getContent(client, url)
		}
		bodys[0] = reduce(bodys[0], bodys[1])
		bodys[2] = reduce(bodys[2], bodys[3])
		web.content = reduce(bodys[0], bodys[2])
		bodyUpdate = true
	}
	title := getTitle(bodys[0])
	if (title != web.Title) {
		web.Title = title
		titleUpdate = true
	}
	return bodyUpdate || titleUpdate
}

func compare(source, newComing string) bool {
	if source == "" { return false }
	checkStrs := strings.Split(source, string(rune(1)))
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>)")
	newComing = string(re.ReplaceAll([]byte(strings.ReplaceAll(newComing, "\n", "nn")), []byte("<script/>")))
	for _, check := range checkStrs {
		if !strings.Contains(newComing, check) {
			fmt.Println(newComing, "\n\n\n", check)
			return false
		}
	}
	return true
}

func substr(s string, start, end int) string {
	if start > len(s) { return "" }
	if end > len(s) {
		if end - start > len(s) { return s }
		return s[start:len(s)]
	}
	return s[start:end]
}

func reduce(b1, b2 string) string {
	re := regexp.MustCompile("(<script.*?/script>|<style.*?/style>)")
	b1 = string(re.ReplaceAll([]byte(strings.ReplaceAll(b1, "\n", "nn")), []byte("<script/>")))
	b2 = string(re.ReplaceAll([]byte(strings.ReplaceAll(b2, "\n", "nn")), []byte("<script/>")))
	sep := string(rune(1))
	result := ""
	groupLen := 30
	MaxGroupDistance := 5
	maxLen := len(b1)
	if len(b2) > len(b1) { maxLen = len(b2) }
	
	i1, i2 := 0, 0

	for i1 < maxLen {
		if b1[i1] == b2[i2] {
			result += string(b1[i1])
		} else {
			if len(result) == 0 || string(result[len(result) - 1]) != sep { result += sep }
			for j := 0; j < MaxGroupDistance; j++ {
				if substr(b1, i1, i1+groupLen) == substr(b2, i2+j, i2+groupLen+j) {
					i2 += j
					result += string(b1[i1])
					break
				}
			}
		}
		i1++
		i2++
		if i1 >= len(b1) || i2 >= len(b2) { break }
	}
	temp := strings.Split(result, sep)
	for i, _ := range(temp) {
		if len(temp[i]) < groupLen {
			temp[i] = ""
		} else {
			temp[i] = temp[i][5:len(temp[i])-5]
		}
	}
	return strings.Join(temp, sep)
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
		web.UpdateTime = time.Now()
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