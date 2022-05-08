package website

import (
	"time"
	"errors"
	"strings"
	"database/sql"
	"net/url"
)

var NotFoundError = errors.New("website not found")

type Website struct {
	UserUUID string
	URL, Title, GroupName string
	content string
	UpdateTime, AccessTime time.Time
}

func NewWebsite(url, userUUID string) Website {
	w := Website{
		URL: url,
		UserUUID: userUUID,
		AccessTime: time.Now(),
		UpdateTime: time.Now(),
	}
	w.Update()
	return w
}

const createSQL = `INSERT OR IGNORE INTO websites 
(url, title, content, update_time) VALUES (?, ?, ?, ?);
INSERT OR IGNORE INTO user_websites
(user_uuid, url, access_time, group_name) VALUES (?, ?, ?, ?);`

func (w Website) Create(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		createSQL, w.URL, w.Title, w.content, w.UpdateTime,
		w.UserUUID, w.URL, w.AccessTime, w.GroupName,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

const updateSQL = `UPDATE websites set 
title=?, content=?, update_time=? where url=?;
UPDATE user_websites SET access_time=?, group_name=? 
where user_uuid=? and url=?`

func (w Website) Save(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		updateSQL, w.Title, w.content, w.UpdateTime, w.URL,
		w.AccessTime, w.GroupName, w.UserUUID, w.URL, 
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (w Website) Delete(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`delete from user_websites where url=? and user_uuid=?;
		delete from websites where url=? and (select count(*) from user_websites where url=?) = 0;`,
		w.URL, w.UserUUID, w.URL, w.URL,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func FindAllWebsites(db *sql.DB) ([]Website, error) {
	rows, err := db.Query("select url, title, content, update_time from websites")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Website, 0)
	for rows.Next() {
		var w Website
		rows.Scan(&w.URL, &w.Title, &w.content, &w.UpdateTime)
		result = append(result, w)
	}
	return result, nil
}

func FindAllUserWebsites(db *sql.DB, userUUID string) ([]Website, error) {
	rows, err := db.Query(
		`select websites.url, title, content, update_time, user_uuid, access_time, group_name
		from websites join user_websites on websites.url=user_websites.url 
		where user_uuid=?`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Website, 0)
	for rows.Next() {
		var w Website
		rows.Scan(
			&w.URL, &w.Title, &w.content, &w.UpdateTime,
			&w.UserUUID, &w.AccessTime, &w.GroupName,
		)
		result = append(result, w)
	}
	return result, nil
}

func FindUserWebsite(db *sql.DB, userUUID, url string) (Website, error) {
	rows, err := db.Query(
		`select websites.url, title, content, update_time, user_uuid, access_time, group_name
		from websites join user_websites on websites.url=user_websites.url 
		where user_uuid=? and websites.url=?`,
		userUUID, url,
	)
	if err != nil {
		return Website{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var w Website
		rows.Scan(
			&w.URL, &w.Title, &w.content, &w.UpdateTime,
			&w.UserUUID, &w.AccessTime, &w.GroupName,
		)
		return w, nil
	}
	return Website{}, NotFoundError
}

func (w Website) Map() map[string]interface{} {
	return map[string]interface{} {
		"url": w.URL,
		"title": w.Title,
		"groupName": w.GroupName,
		"updateTime": w.UpdateTime,
		"accessTime": w.AccessTime,
	}
}

func (w Website) Host() string {
	u, err := url.Parse(w.URL)
	if err != nil || w.URL == "" {
		return ""
	}
	host := u.Host
	splitedHost := strings.Split(host, ".")
	return strings.Join(splitedHost[len(splitedHost) - 2:], ".")
}