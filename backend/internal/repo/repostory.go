package repo

import (
	"github.com/htchan/WebHistory/internal/model"
	"database/sql"
)

type Repostory interface {
	CreateWebsite(*model.Website) error
	UpdateWebsite(*model.Website) error
	DeleteWebsite(*model.Website) error

	FindWebsites() ([]model.Website, error)
	FindWebsite(uuid string) (*model.Website, error)

	CreateUserWebsite(*model.UserWebsite) error
	UpdateUserWebsite(*model.UserWebsite) error
	DeleteUserWebsite(*model.UserWebsite) error

	FindUserWebsites(userUUID string) (model.UserWebsites, error)
	FindUserWebsitesByGroup(userUUID, group string) (model.WebsiteGroup, error)
	FindUserWebsite(userUUID, websiteUUID string) (*model.UserWebsite, error)

	Stats() (sql.DBStats)
}
