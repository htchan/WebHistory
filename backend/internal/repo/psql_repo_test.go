package repo

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/model"
	_ "github.com/lib/pq"
)

func openPsql(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=test password=test dbname=test sslmode=disable")
	if err != nil {
		t.Fatalf("got error when open database: %v", err)
		return nil
	}
	return db
}

func populateData(db *sql.DB) error {
	_, err := db.Exec("insert into websites (uuid, url, title, content, update_time) values ('abc', 'http://example.com', 'title', 'content', '2000-01-02 03:04:05-06')")
	if err != nil {
		return err
	}
	_, err = db.Exec("insert into user_websites (website_uuid, user_uuid, group_name, access_time) values ('abc', 'def', 'title', '2000-01-02 03:04:05-06')")
	return err
}

func TestNewPsqlRepo(t *testing.T) {
	db := openPsql(t)
	defer db.Close()

	tests := []struct {
		name string
		db   *sql.DB
	}{
		{
			name: "providing a psql database",
			db:   db,
		},
		{
			name: "providing nil database",
			db:   nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			repo := NewPsqlRepo(test.db, &config.Config{})
			if repo.db != test.db {
				t.Errorf("db in repo is different from when provided")
				t.Error(repo.db)
				t.Error(test.db)
			}
		})
	}
}

func TestPsqlRepo_CreateWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		web       model.Website
		expect    model.Website
		expectErr bool
	}{
		{
			name: "create a new website",
			web: model.Website{
				UUID:       "dcb12928-5b5b-43f3-9d0e-ddb526d9794d",
				URL:        "http://example.com",
				Title:      "unknown",
				RawContent: "",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: model.Website{
				UUID:       "dcb12928-5b5b-43f3-9d0e-ddb526d9794d",
				URL:        "http://example.com",
				Title:      "unknown",
				RawContent: "",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectErr: false,
		},
		{
			name: "create an existing website",
			web: model.Website{
				UUID:       "f9707295-e5cd-4651-a82d-595ac8022eea",
				URL:        "http://example.com",
				Title:      "",
				RawContent: "",
				UpdateTime: time.Now(),
			},
			expect: model.Website{
				UUID:       "dcb12928-5b5b-43f3-9d0e-ddb526d9794d",
				URL:        "http://example.com",
				Title:      "unknown",
				RawContent: "",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := r.CreateWebsite(&test.web)
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(test.web, test.expect) {
				t.Errorf("result web different from expect web")
				t.Error(test.web)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_UpdateWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)

	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		web       model.Website
		expect    *model.Website
		expectErr bool
	}{
		{
			name: "update successfully",
			web: model.Website{
				UUID:       "abc",
				URL:        "http://example.com",
				Title:      "unknown",
				RawContent: "content new",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: &model.Website{
				UUID:       "abc",
				URL:        "http://example.com",
				Title:      "unknown",
				RawContent: "content new",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectErr: false,
		},
		{
			name: "update not exist website",
			web: model.Website{
				UUID:       "dcb12928-5b5b-43f3-9d0e-ddb526d9794d",
				URL:        "http://example.com",
				Title:      "unknown",
				RawContent: "",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect:    nil,
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := r.UpdateWebsite(&test.web)
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			web, _ := r.FindWebsite(test.web.UUID)
			if !cmp.Equal(web, test.expect) {
				t.Errorf("web in database different from expected")
				t.Error(web)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_DeleteWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		webUUID   string
		expectErr bool
	}{
		{
			name:      "delete successfully",
			webUUID:   "abc",
			expectErr: false,
		},
		{
			name:      "delete not exist",
			webUUID:   "not exist",
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := r.DeleteWebsite(&model.Website{UUID: test.webUUID})
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			web, err := r.FindWebsite(test.webUUID)
			if err == nil || web != nil {
				t.Errorf("got error: %v", err)
				t.Errorf("find website: %v", web)
			}
		})
	}
}

func TestPsqlRepo_FindWebsites(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		expect    []model.Website
		expectErr bool
	}{
		{
			name: "happy flow",
			expect: []model.Website{
				{
					UUID:       "abc",
					URL:        "http://example.com",
					Title:      "title",
					RawContent: "content",
					UpdateTime: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := r.FindWebsites()
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(result, test.expect) {
				t.Errorf("result different from expected")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_FindWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		webUUID   string
		expect    *model.Website
		expectErr bool
	}{
		{
			name:    "find exist website",
			webUUID: "abc",
			expect: &model.Website{
				UUID:       "abc",
				URL:        "http://example.com",
				Title:      "title",
				RawContent: "content",
				UpdateTime: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
			},
			expectErr: false,
		},
		{
			name:      "find not exist website",
			webUUID:   "unknown",
			expect:    nil,
			expectErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := r.FindWebsite(test.webUUID)
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(result, test.expect) {
				t.Errorf("result different from expected")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_CreateUserWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})
	tests := []struct {
		name      string
		web       model.UserWebsite
		expect    model.UserWebsite
		expectErr bool
	}{
		{
			name: "create new user website",
			web: model.UserWebsite{
				WebsiteUUID: "abc",
				UserUUID:    "new",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: model.UserWebsite{
				WebsiteUUID: "abc",
				UserUUID:    "new",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Website: model.Website{
					UUID:       "abc",
					URL:        "http://example.com",
					Title:      "title",
					UpdateTime: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			expectErr: false,
		},
		{
			name: "create existing user website",
			web: model.UserWebsite{
				WebsiteUUID: "abc",
				UserUUID:    "new",
			},
			expect: model.UserWebsite{
				WebsiteUUID: "abc",
				UserUUID:    "new",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Website: model.Website{
					UUID:       "abc",
					URL:        "http://example.com",
					Title:      "title",
					UpdateTime: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			expectErr: false,
		},
		{
			name: "create new user website link to not exist website",
			web: model.UserWebsite{
				WebsiteUUID: "new",
				UserUUID:    "new",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: model.UserWebsite{
				WebsiteUUID: "new",
				UserUUID:    "new",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := r.CreateUserWebsite(&test.web)
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(test.web, test.expect) {
				t.Errorf("result different from expect")
				t.Error(test.web)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_UpdteUserWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})
	tests := []struct {
		name      string
		web       model.UserWebsite
		expect    *model.UserWebsite
		expectErr bool
	}{
		{
			name: "update existing website",
		},
		{
			name: "update not exist user website",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := r.UpdateUserWebsite(&test.web)
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			web, _ := r.FindUserWebsite(test.web.UserUUID, test.web.WebsiteUUID)
			if !cmp.Equal(web, test.expect) {
				t.Errorf("web in database different from expected")
				t.Error(web)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_DeleteUserWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		userUUID  string
		webUUID   string
		expectErr bool
	}{
		{
			name:      "delete successfully",
			webUUID:   "abc",
			userUUID:  "def",
			expectErr: false,
		},
		{
			name:      "delete not exist",
			webUUID:   "not exist",
			userUUID:  "not exist",
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			err := r.DeleteUserWebsite(&model.UserWebsite{
				UserUUID:    test.userUUID,
				WebsiteUUID: test.webUUID,
			})
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			web, err := r.FindUserWebsite(test.userUUID, test.webUUID)
			if err == nil || web != nil {
				t.Errorf("got error: %v", err)
				t.Errorf("find website: %v", web)
			}
		})
	}
}

func TestPsqlRepo_FindUserWebsites(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		userUUID  string
		expect    model.UserWebsites
		expectErr bool
	}{
		{
			name:     "find web of existing user",
			userUUID: "def",
			expect: model.UserWebsites{
				{
					UserUUID:    "def",
					WebsiteUUID: "abc",
					GroupName:   "title",
					AccessTime:  time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
					Website: model.Website{
						UUID:       "abc",
						URL:        "http://example.com",
						Title:      "title",
						UpdateTime: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "find web of not existing user",
			userUUID:  "not exist",
			expect:    nil,
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := r.FindUserWebsites(test.userUUID)

			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(result, test.expect) {
				t.Errorf("result different from expected")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_FindUserWebsitesByGroup(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		userUUID  string
		group     string
		expect    model.WebsiteGroup
		expectErr bool
	}{
		{
			name:     "find web of existing group and user",
			userUUID: "def",
			group:    "title",
			expect: model.WebsiteGroup{
				{
					UserUUID:    "def",
					WebsiteUUID: "abc",
					GroupName:   "title",
					AccessTime:  time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
					Website: model.Website{
						UUID:       "abc",
						URL:        "http://example.com",
						Title:      "title",
						UpdateTime: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "find web of not existing group",
			userUUID:  "def",
			group:     "not exist",
			expect:    nil,
			expectErr: false,
		},
		{
			name:      "find web of not existing user",
			userUUID:  "not exist",
			group:     "title",
			expect:    nil,
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := r.FindUserWebsitesByGroup(test.userUUID, test.group)

			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(result, test.expect) {
				t.Errorf("result different from expected")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

func TestPsqlRepo_FindUserWebsite(t *testing.T) {
	db := openPsql(t)
	r := NewPsqlRepo(db, &config.Config{})
	populateData(db)
	t.Cleanup(func() {
		db.Exec("delete from websites")
		db.Exec("delete from user_websites")
		db.Close()
	})

	tests := []struct {
		name      string
		userUUID  string
		webUUID   string
		expect    *model.UserWebsite
		expectErr bool
	}{
		{
			name:     "find web of existing group and user",
			userUUID: "def",
			webUUID:  "abc",
			expect: &model.UserWebsite{
				UserUUID:    "def",
				WebsiteUUID: "abc",
				GroupName:   "title",
				AccessTime:  time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
				Website: model.Website{
					UUID:       "abc",
					URL:        "http://example.com",
					Title:      "title",
					UpdateTime: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
				},
			},
			expectErr: false,
		},
		{
			name:      "find web of not existing web uuid",
			userUUID:  "def",
			webUUID:   "not exist",
			expect:    nil,
			expectErr: true,
		},
		{
			name:      "find web of not existing user",
			userUUID:  "not exist",
			webUUID:   "abc",
			expect:    nil,
			expectErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := r.FindUserWebsite(test.userUUID, test.webUUID)

			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(result, test.expect) {
				t.Errorf("result different from expected")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}
