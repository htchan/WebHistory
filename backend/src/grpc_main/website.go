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

	pb "github.com/htchan/WebHistory/src/protobuf"
)

var database *sql.DB

type Website struct {
	Url string
	content string
	UpdateTime, AccessTime int64
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
		"select url, content, updateTime, accessTime from websites where url=?", url)
	if err != nil { panic(err) }
	var web Website
	if rows.Next() {
		rows.Scan(&web.Url, &web.content, &web.UpdateTime, &web.AccessTime)
	}
	err = rows.Close()
	if err != nil { panic(err) }
	return web
}

func (web *Website) _checkTimeUpdate(timeStr string) bool {
	if timeStr == "" { return false }
	layout := "Mon, 2 Jan 2006 15:04:05 GMT"
	t, err := time.Parse(layout, timeStr)
	if err == nil && t.After(time.Unix(int64(web.UpdateTime), 0)) {
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

func (web *Website) _checkBodyUpdate(client http.Client, url string) bool {
	bodys := make([]string, 4)
	bodys[0] = getContent(client, url)
	if !compare(web.content, bodys[0]) {
		for i := 1; i < 4; i++ {
			bodys[i] = getContent(client, url)
		}
		bodys[0] = reduce(bodys[0], bodys[1])
		bodys[2] = reduce(bodys[2], bodys[3])
		web.content = reduce(bodys[0], bodys[2])
		return true
	}
	return false
}

func compare(source, newComing string) bool {
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
			if string(result[len(result) - 1]) != sep { result += sep }
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
	if err != nil { panic(err) }
	if web._checkTimeUpdate(resp.Header.Get("last-modified")) ||
		web._checkBodyUpdate(client, web.Url) {
		web.UpdateTime = time.Now().Unix()
	}
}

func (web Website) insert(tx *sql.Tx) {
	_, err := tx.Exec("insert into websites (url, content, updateTime, accessTime) values (?, ?, ?, ?)",
		web.Url, web.content, web.UpdateTime, web.AccessTime)
	if err != nil { panic(err) }
}

func (web Website) Save() {
	tx, err := database.Begin()
	if err != nil { panic(err) }
	result, err := tx.Exec("update websites set content=?, updateTime=?, accessTime=? where url=?",
		web.content, web.UpdateTime, web.AccessTime, web.Url)
	if err != nil { panic(err) }
	rowsAffected, err := result.RowsAffected()
	if err != nil { panic(err) }
	if rowsAffected == 0 { web.insert(tx) }
	err = tx.Commit()
	if err != nil { panic(err) }
}

func (web Website) Response() *pb.WebsiteResponse {
	return &pb.WebsiteResponse{
		Url: web.Url,
		UpdateTime: web.UpdateTime,
		AccessTime: web.AccessTime,
	}
}