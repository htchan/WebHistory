package website

import (
	"time"
	"errors"
	"strings"
	"database/sql"
	"net/url"
	"encoding/json"
	"github.com/google/uuid"
)

var NotFoundError = errors.New("website not found")

type Website struct {
	UUID string
	UserUUID string
	URL, Title, GroupName string
	content string
	UpdateTime, AccessTime time.Time
}

func NewWebsite(url, userUUID string) Website {
	web := Website{
		UUID: uuid.New().String(),
		URL: url,
		UserUUID: userUUID,
		AccessTime: time.Now(),
		UpdateTime: time.Now(),
	}
	web.Update()
	return web
}

const createWebsiteSQL = `INSERT OR IGNORE INTO websites 
(uuid, url, title, content, update_time) VALUES (?, ?, ?, ?, ?);`

const createUserWebsiteSQL = `INSERT OR IGNORE INTO user_websites
(uuid, user_uuid, access_time, group_name) VALUES
((select uuid from websites where url=?), ?, ?, ?) returning uuid;`

func (web *Website) Create(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		createWebsiteSQL, web.UUID, web.URL, web.Title, web.content, web.UpdateTime)
	if err != nil {
		tx.Rollback()
		return err
	}
	rows, err := tx.Query(
		createUserWebsiteSQL, web.URL, web.UserUUID, web.AccessTime, web.GroupName,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows.Next() {
		rows.Scan(&web.UUID)
	}
	return tx.Commit()
}

const updateSQL = `UPDATE websites set 
title=?, content=?, update_time=? where UUID=?;
UPDATE user_websites SET access_time=?, group_name=? 
where user_uuid=? and UUID=?`

func (web Website) Save(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		updateSQL, web.Title, web.content, web.UpdateTime, web.UUID,
		web.AccessTime, web.GroupName, web.UserUUID, web.UUID, 
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (web Website) Delete(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`delete from user_websites where uuid=? and user_uuid=?;
		delete from websites where uuid=? and (select count(*) from user_websites where uuid=?) = 0;`,
		web.UUID, web.UserUUID, web.UUID, web.UUID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func FindAllWebsites(db *sql.DB) ([]Website, error) {
	rows, err := db.Query("select uuid, url, title, content, update_time from websites")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Website, 0)
	for rows.Next() {
		var web Website
		rows.Scan(&web.UUID, &web.URL, &web.Title, &web.content, &web.UpdateTime)
		result = append(result, web)
	}
	return result, nil
}

func FindAllUserWebsites(db *sql.DB, userUUID string) ([]Website, error) {
	rows, err := db.Query(
		`select websites.uuid, url, title, content, update_time, user_uuid, access_time, group_name
		from websites join user_websites on websites.uuid=user_websites.uuid 
		where user_uuid=? order by (update_time > access_time) desc, update_time desc, access_time desc`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Website, 0)
	for rows.Next() {
		var web Website
		rows.Scan(
			&web.UUID, &web.URL, &web.Title, &web.content, &web.UpdateTime,
			&web.UserUUID, &web.AccessTime, &web.GroupName,
		)
		result = append(result, web)
	}
	return result, nil
}

func FindUserWebsite(db *sql.DB, userUUID, webUUID string) (Website, error) {
	rows, err := db.Query(
		`select websites.uuid, url, title, content, update_time, user_uuid, access_time, group_name
		from websites join user_websites on websites.uuid=user_websites.uuid 
		where user_uuid=? and websites.uuid=?`,
		userUUID, webUUID,
	)
	if err != nil {
		return Website{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var web Website
		rows.Scan(
			&web.UUID, &web.URL, &web.Title, &web.content, &web.UpdateTime,
			&web.UserUUID, &web.AccessTime, &web.GroupName,
		)
		return web, nil
	}
	return Website{}, NotFoundError
}

func (web Website) Map() map[string]interface{} {
	return map[string]interface{} {
		"uuid": web.UUID,
		"url": web.URL,
		"title": web.Title,
		"groupName": web.GroupName,
		"updateTime": web.UpdateTime,
		"accessTime": web.AccessTime,
	}
}

func (web Website) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UUID string `json:"uuid"`
		URL string `json:"url"`
		Title string `json:"title"`
		GroupName string `json:"group_name"`
		UpdateTime time.Time `json:"update_time"`
		AccessTime time.Time `json:"access_time"`
	}{
		UUID: web.UUID,
		URL: web.URL,
		Title: web.Title,
		GroupName: web.GroupName,
		UpdateTime: web.UpdateTime,
		AccessTime: web.AccessTime,
	})
}

func (web Website) Host() string {
	u, err := url.Parse(web.URL)
	if err != nil || web.URL == "" {
		return ""
	}
	host := u.Host
	splitedHost := strings.Split(host, ".")
	return strings.Join(splitedHost[len(splitedHost) - 2:], ".")
}

type WebsiteGroup []Website

func WebsitesToWebsiteGroups(websites []Website) []WebsiteGroup {
	websiteGroupMap := make(map[string]WebsiteGroup)
	for _, web := range websites {
		group, _ := websiteGroupMap[web.GroupName]
		websiteGroupMap[web.GroupName] = append(group, web)
	}
	result := make([]WebsiteGroup, len(websiteGroupMap))
	i := 0
	for _, item := range websiteGroupMap {
		result[i] = item
		i++
	}
	return result
}