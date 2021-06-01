package main

import (
	"time"
	"net/http"
	"io/ioutil"
	
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

type Website struct {
	Url string
	content string
	UpdateTime time.Time
	Updated bool
}

func openDatabase(location string) {
	database, err := sql.Open("sqlite3", location)
	if err != nil { panic(err) }
	database.SetMaxIdleConns(5);
	database.SetMaxOpenConns(50);
}

func closeDatabase() {
	database.Close()
}

func Urls() []string {
	result := make([]string, 0)
	rows, err := database.Query("select url from website")
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
		"select url, content, updateTime, update from website where url=?", url)
	if err != nil { panic(err) }
	var web Website
	var t int
	if rows.Next() {
		rows.Scan(&web.Url, &web.content, &t, &web.Updated)
		web.UpdateTime = time.Unix(int64(t), 0)
	}
	return web
}

func (web *Website) _checkTimeUpdate(timeStr string) bool {
	layout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(layout, timeStr)
	if err == nil && t.After(web.UpdateTime) {
		return true
	}
	return false
}

func (web *Website) _checkBodyUpdate(body string) bool {
	if web.content != body {
		web.content = body
		return true
	}
	return false
}

func (web *Website) update() {
	if web.Updated == true { return }
	client := http.Client{Timeout: 30*time.Second}
	resp, err := client.Get(web.Url);
	if err != nil { return }
	bodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil { return }
	if web._checkTimeUpdate(resp.Header.Get("last-modified")) ||
		web._checkBodyUpdate(string(bodyByte)) {
		web.UpdateTime = time.Now()
		web.Updated = true
	}
}

func (web Website) Update() {
	web.update()
	if web.Updated == false { return }
	tx, err := database.Begin()
	if err != nil { panic(err) }
	_, err = tx.Exec("update website set content=?, updateTime=?, update=? where url=?",
		web.content, web.UpdateTime.Unix(), web.Update, web.Url)
	if err != nil { panic(err) }
	err = tx.Commit()
	if err != nil { panic(err) }
}

func (web Website) Map() map[string]interface{} {
	return map[string]interface{} {
		"url": web.Url,
		"content": web.content,
		"updateTime": web.UpdateTime,
	}
}