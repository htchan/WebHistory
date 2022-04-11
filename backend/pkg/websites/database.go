package websites

import (
	"time"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/htchan/WebHistory/internal/logging"
)

var database *sql.DB

func OpenDatabase(location string) {
	var err error
	database, err = sql.Open("sqlite3", location)
	if err != nil { panic(err) }
	database.SetMaxIdleConns(5);
	database.SetMaxOpenConns(50);
	logging.Log("database.open", database)
}

func closeDatabase() {
	logging.Log("dataabse.close", database)
	database.Close()
}

func FindAllWebsites() []Website {
	resultUpdate := make([]Website, 0)
	resultUnchange := make([]Website, 0)
	rows, err := database.Query("select user_uuid, url, updateTime, accessTime from websites order by groupName, updateTime desc")
	if err != nil { panic(err) }
	var tempUserUUID, tempUrl string
	var updateTime, accessTime int64
	for rows.Next() {
		rows.Scan(&tempUserUUID, &tempUrl, &updateTime, &accessTime)
		if updateTime > accessTime {
			resultUpdate = append(resultUpdate, FindWebsiteByUrl(tempUserUUID, tempUrl))
		} else {
			resultUnchange = append(resultUnchange, FindWebsiteByUrl(tempUserUUID, tempUrl))
		}
	}
	return append(resultUpdate, resultUnchange...)
}

func FindUrlsByUser(userUUID string) []string {
	resultUpdate := make([]string, 0)
	resultUnchange := make([]string, 0)
	rows, err := database.Query("select url, updateTime, accessTime from websites where user_uuid=? order by groupName, updateTime desc", userUUID)
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

func FindAllGroupNames(userUUID string) []string {
	result := make([]string, 0)
	rows, err := database.Query("select groupName from websites where user_uuid=? group by groupName order by max(updateTime) desc", userUUID)
	if err != nil { panic(err) }
	var temp string
	for rows.Next() {
		rows.Scan(&temp)
		result = append(result, temp)
	}
	return result
}

func FindUrlsByGroupName(userUUID, groupName string) []string {
	urls := make([]string, 0)
	rows, err := database.Query("select url from websites where user_uuid=? and groupName=? order by updateTime desc", userUUID, groupName)
	if err != nil { panic(err) }
	var temp string
	for rows.Next() {
		rows.Scan(&temp)
		urls = append(urls, temp)
	}
	return urls
}

func FindWebsiteByUrl(userUUID, url string) Website {
	rows, err := database.Query(
		"select user_uuid,url, title, groupName, content, updateTime, accessTime from websites " +
		"where user_uuid=? and url=?", userUUID, url)
	if err != nil { panic(err) }
	var web Website
	var updateTime, accessTime int
	if rows.Next() {
		rows.Scan(&web.UserUUID, &web.Url, &web.Title, &web.GroupName, &web.content, &updateTime, &accessTime)
		web.UpdateTime = time.Unix(int64(updateTime), 0)
		web.AccessTime = time.Unix(int64(accessTime), 0)
	}
	err = rows.Close()
	if err != nil { panic(err) }
	return web
}

func (website Website) create(tx *sql.Tx) {
	_, err := tx.Exec("insert into websites (user_uuid, url, title, groupName, content, updateTime, accessTime) " +
		"values (?, ?, ?, ?, ?, ?, ?)",
		website.UserUUID, website.Url, website.Title, website.GroupName, website.content, website.UpdateTime.Unix(), website.AccessTime.Unix())
	if err != nil { panic(err) }
}

func (website Website) Save() {
	tx, err := database.Begin()
	if err != nil { panic(err) }
	result, err := tx.Exec("update websites set " +
		"title=?, groupName=?, content=?, updateTime=?, accessTime=? where user_uuid=? and url=?",
		website.Title, website.GroupName, website.content, 
		website.UpdateTime.Unix(), website.AccessTime.Unix(),
		website.UserUUID, website.Url)
	if err != nil { panic(err) }
	rowsAffected, err := result.RowsAffected()
	if err != nil { panic(err) }
	if rowsAffected == 0 { website.create(tx) }
	err = tx.Commit()
	if err != nil { panic(err) }
}

func (website Website) Delete() {
	tx, err := database.Begin()
	if err != nil { panic(err) }
	_, err = tx.Exec("delete from websites where user_uuid=? and url=?", website.UserUUID, website.Url)
	if err != nil { panic(err) }
	err = tx.Commit()
	if err != nil { panic(err) }

}