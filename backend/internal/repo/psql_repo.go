package repo

import (
	"database/sql"
	"fmt"

	"github.com/htchan/WebHistory/internal/model"
)

type PsqlRepo struct {
	db *sql.DB
}

var _ Repostory = &PsqlRepo{}

func NewPsqlRepo(db *sql.DB) *PsqlRepo {
	return &PsqlRepo{db: db}
}

func (r *PsqlRepo) CreateWebsite(web *model.Website) error {
	// return web if url exist
	rows, err := r.db.Query("select uuid, url, title, content, update_time from websites where url=$1;", web.URL)
	if err == nil && rows.Next() {
		err = rows.Scan(&web.UUID, &web.URL, &web.Title, &web.RawContent, &web.UpdateTime)
		if err == nil {
			return nil
		}
	}
	// create website if url not exist
	_, err = r.db.Exec(
		"insert into websites (uuid, url, title, content, update_time) values ($1, $2, $3, $4, $5);",
		web.UUID, web.URL, web.Title, web.RawContent, web.UpdateTime,
	)
	if err != nil {
		return fmt.Errorf("fail to insert website: %w", err)
	}

	return nil
}

func (r *PsqlRepo) UpdateWebsite(web *model.Website) error {
	_, err := r.db.Exec(
		"UPDATE websites set url=$1, title=$2, content=$3, update_time=$4 where uuid=$5",
		web.URL, web.Title, web.RawContent, web.UpdateTime, web.UUID,
	)
	if err != nil {
		return fmt.Errorf("fail to update website: %w", err)
	}
	return nil
}

func (r *PsqlRepo) DeleteWebsite(web *model.Website) error {
	_, err := r.db.Exec("delete from websites where uuid=$1", web.UUID)
	if err != nil {
		return fmt.Errorf("fail to delete website: %w", err)
	}
	return nil
}

func (r *PsqlRepo) FindWebsites() ([]model.Website, error) {
	rows, err := r.db.Query("select uuid, url, title, content, update_time from websites")
	if err != nil {
		return nil, fmt.Errorf("fail to fetch websites: %w", err)
	}

	var webs []model.Website

	for rows.Next() {
		var web model.Website
		err := rows.Scan(&web.UUID, &web.URL, &web.Title, &web.RawContent, &web.UpdateTime)
		if err != nil {
			return webs, fmt.Errorf("fail to read websites: %w", err)
		}
		webs = append(webs, web)
	}

	return webs, nil
}

func (r *PsqlRepo) FindWebsite(uuid string) (*model.Website, error) {
	rows, err := r.db.Query("select uuid, url, title, content, update_time from websites where uuid=$1", uuid)
	if err != nil {
		return nil, fmt.Errorf("fail to fetch website: %w", err)
	}

	if rows.Next() {
		web := new(model.Website)
		err := rows.Scan(&web.UUID, &web.URL, &web.Title, &web.RawContent, &web.UpdateTime)
		if err != nil {
			return nil, fmt.Errorf("fail to read websites: %w", err)
		}
		return web, nil
	}

	return nil, fmt.Errorf("fail to fetch website: website not found")
}

func (r *PsqlRepo) CreateUserWebsite(web *model.UserWebsite) error {
	rows, err := r.db.Query(
		`select 
		user_uuid, website_uuid, access_time, group_name, 
		uuid, url, title, update_time 
		from user_websites join websites on user_websites.website_uuid=websites.uuid 
		where user_uuid=$1 and website_uuid=$2;`,
		web.UserUUID, web.WebsiteUUID,
	)
	if err == nil && rows.Next() {
		err = rows.Scan(
			&web.UserUUID, &web.WebsiteUUID, &web.AccessTime, &web.GroupName,
			&web.Website.UUID, &web.Website.URL, &web.Website.Title, &web.Website.UpdateTime,
		)
		return err
	}

	w, err := r.FindWebsite(web.WebsiteUUID)
	if err != nil {
		return fmt.Errorf("fail to create user website: %w", err)
	}
	web.Website = *w

	_, err = r.db.Exec(
		`insert into user_websites 
		(user_uuid, website_uuid, access_time, group_name) 
		values 
		($1, $2, $3, $4);`,
		web.UserUUID, web.WebsiteUUID, web.AccessTime, web.GroupName,
	)
	if err != nil {
		return fmt.Errorf("fail to create user website: %w", err)
	}

	return nil
}

func (r *PsqlRepo) UpdateUserWebsite(web *model.UserWebsite) error {
	_, err := r.db.Exec(
		`update user_websites 
		set access_time=$3, group_name=$4 
		where user_uuid=$2 and website_uuid=$1`,
		web.UserUUID, web.WebsiteUUID, web.AccessTime, web.GroupName,
	)
	if err != nil {
		return fmt.Errorf("fail to update user website: %w", err)
	}
	return nil
}

func (r *PsqlRepo) DeleteUserWebsite(web *model.UserWebsite) error {
	_, err := r.db.Exec(
		"delete from user_websites where website_uuid=$1 and user_uuid=$2",
		web.WebsiteUUID, web.UserUUID,
	)
	if err != nil {
		return fmt.Errorf("fail to delete user website: %w", err)
	}
	return nil
}

func (r *PsqlRepo) FindUserWebsites(userUUID string) (model.UserWebsites, error) {
	rows, err := r.db.Query(
		`select 
		website_uuid, user_uuid, access_time, group_name ,
		uuid, url, title, update_time 
		from user_websites join websites on user_websites.website_uuid=websites.uuid 
		where user_uuid=$1
		order by (update_time > access_time) desc, update_time desc, access_time desc`,
		userUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("fail to fetch user websites: %w", err)
	}

	var webs model.UserWebsites

	for rows.Next() {
		var web model.UserWebsite

		err := rows.Scan(
			&web.WebsiteUUID, &web.UserUUID, &web.AccessTime, &web.GroupName,
			&web.Website.UUID, &web.Website.URL, &web.Website.Title, &web.Website.UpdateTime,
		)
		if err != nil {
			return webs, fmt.Errorf("fail to read user websites: %w", err)
		}

		webs = append(webs, web)
	}

	return webs, nil
}

func (r *PsqlRepo) FindUserWebsitesByGroup(userUUID, group string) (model.WebsiteGroup, error) {
	rows, err := r.db.Query(
		`select 
		website_uuid, user_uuid, access_time, group_name ,
		uuid, url, title, update_time 
		from user_websites join websites on user_websites.website_uuid=websites.uuid 
		where user_uuid=$1 and group_name=$2`,
		userUUID, group,
	)
	if err != nil {
		return nil, fmt.Errorf("fail to fetch user websites: %w", err)
	}

	var webs model.WebsiteGroup

	for rows.Next() {
		var web model.UserWebsite

		err := rows.Scan(
			&web.WebsiteUUID, &web.UserUUID, &web.AccessTime, &web.GroupName,
			&web.Website.UUID, &web.Website.URL, &web.Website.Title, &web.Website.UpdateTime,
		)
		if err != nil {
			return webs, fmt.Errorf("fail to read user websites: %w", err)
		}

		webs = append(webs, web)
	}

	return webs, nil
}

func (r *PsqlRepo) FindUserWebsite(userUUID, websiteUUID string) (*model.UserWebsite, error) {
	rows, err := r.db.Query(
		`select 
		website_uuid, user_uuid, access_time, group_name ,
		uuid, url, title, update_time 
		from user_websites join websites on user_websites.website_uuid=websites.uuid 
		where user_uuid=$1 and website_uuid=$2`,
		userUUID, websiteUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("fail to fetch user website: %w", err)
	}

	if rows.Next() {
		web := new(model.UserWebsite)

		err := rows.Scan(
			&web.WebsiteUUID, &web.UserUUID, &web.AccessTime, &web.GroupName,
			&web.Website.UUID, &web.Website.URL, &web.Website.Title, &web.Website.UpdateTime,
		)
		if err != nil {
			return web, fmt.Errorf("fail to read user website: %w", err)
		}

		return web, nil
	}

	return nil, fmt.Errorf("fail to find user website")
}
