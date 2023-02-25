package psql

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/go-cmp/cmp"
	"github.com/htchan/WebHistory/internal/config"
	"github.com/htchan/WebHistory/internal/model"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

var connString string

const (
	user     = "test"
	password = "test"
	dbname   = "test"
)

func setupMigrate(connString string) error {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return fmt.Errorf("migrate fail: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file:///home/htchan/Project/WebHistory/backend/migrations", "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		return err
	}

	defer m.Close()

	return nil
}

func setupContainer() (string, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", func() {}, fmt.Errorf("init docker fail: %w", err)
	}

	containerName := "webhistory_test_psql_db"
	pool.RemoveContainerByName(containerName)

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "latest",
			Name:       containerName,
			Env: []string{
				fmt.Sprintf("POSTGRES_USER=%s", user),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
				fmt.Sprintf("POSTGRES_DB=%s", dbname),
			},
		},
		func(hc *docker.HostConfig) {
			hc.AutoRemove = true
			hc.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		return "", func() {}, fmt.Errorf("create resource fail: %w", err)
	}

	purge := func() {
		err := resource.Close()
		if err != nil {
			fmt.Println("purge error", err)
		}
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, resource.GetHostPort("5432/tcp"), dbname)
	time.Sleep(3 * time.Second)

	err = setupMigrate(connString)
	if err != nil {
		return "", purge, fmt.Errorf("migrate fail: %w", err)
	}

	return connString, purge, nil
}

func TestMain(m *testing.M) {
	psqlConnString, purge, err := setupContainer()
	if err != nil {
		log.Fatalf("fail to setup docker: %v", err)
	}

	connString = psqlConnString

	goleak.VerifyTestMain(m)

	purge()
}

func populateData(db *sql.DB, uuid, title string) error {
	_, err := db.Exec("insert into websites (uuid, url, title, content, update_time) values ($1, $2, $3, 'content', $4)", uuid, "http://example.com/"+title, title, time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC))
	if err != nil {
		return err
	}
	_, err = db.Exec("insert into user_websites (website_uuid, user_uuid, group_name, access_time) values ($1, 'def', $2, $3)", uuid, title, time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC))
	return err
}

func TestNewRepo(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

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
			t.Parallel()

			repo := NewRepo(test.db, &config.Config{})
			assert.Equal(t, test.db, repo.db)
		})
	}
}

func TestPsqlRepo_CreateWebsite(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})
	t.Cleanup(func() {
		db.Exec("delete from websites where title=$1", "unknown")
		db.Close()
	})

	uuid := "create-website-uuid"
	title := "create website"
	populateData(db, uuid, title)

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
				UUID:       uuid,
				URL:        "http://example.com/" + title,
				Title:      title,
				RawContent: "",
				UpdateTime: time.Now().UTC(),
			},
			expect: model.Website{
				UUID:       uuid,
				URL:        "http://example.com/" + title,
				Title:      title,
				RawContent: "content",
				UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := r.CreateWebsite(&test.web)
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			if !cmp.Equal(test.web, test.expect) {
				t.Errorf("result web different from expect web")
				t.Error(cmp.Diff(test.expect, test.web))
			}
		})
	}
}

func TestPsqlRepo_UpdateWebsite(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "update-website-uuid"
	title := "update website"
	populateData(db, uuid, title)

	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", title)
		db.Exec("delete from user_websites where website_uuid=$1", title)
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
				UUID:       uuid,
				URL:        "http://example.com/" + title,
				Title:      title,
				RawContent: "content new",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: &model.Website{
				UUID:       uuid,
				URL:        "http://example.com/" + title,
				Title:      title,
				RawContent: "content new",
				UpdateTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectErr: false,
		},
		{
			name: "update not exist website",
			web: model.Website{
				UUID:       "uuid-that-not-exist",
				URL:        "http://example.com/not-exist",
				Title:      title,
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
			t.Parallel()

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
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "delete-website-uuid"
	title := "delete website"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
		db.Close()
	})

	tests := []struct {
		name      string
		webUUID   string
		expectErr bool
	}{
		{
			name:      "delete successfully",
			webUUID:   uuid,
			expectErr: false,
		},
		{
			name:      "delete not exist",
			webUUID:   "uuid-that-not-exist",
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "find-websites-uuid"
	title := "find websites"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
		db.Close()
	})

	tests := []struct {
		name      string
		expect    model.Website
		expectErr bool
	}{
		{
			name: "happy flow",
			expect: model.Website{
				UUID:       uuid,
				URL:        "http://example.com/" + title,
				Title:      title,
				RawContent: "content",
				UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := r.FindWebsites()
			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			for _, web := range result {
				if cmp.Equal(web, test.expect) {
					return
				}
			}
			t.Errorf("result not contains expected")
			t.Error(result)
			t.Error(test.expect)
		})
	}
}

func TestPsqlRepo_FindWebsite(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "find-website-uuid"
	title := "find website"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
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
			webUUID: uuid,
			expect: &model.Website{
				UUID:       uuid,
				URL:        "http://example.com/" + title,
				Title:      title,
				RawContent: "content",
				UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			expectErr: false,
		},
		{
			name:      "find not exist website",
			webUUID:   "uuid-that-not-exist",
			expect:    nil,
			expectErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "create-user-website-uuid"
	title := "create user website"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
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
				WebsiteUUID: uuid,
				UserUUID:    "new-user-website-uuid",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: model.UserWebsite{
				WebsiteUUID: uuid,
				UserUUID:    "new-user-website-uuid",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Website: model.Website{
					UUID:       uuid,
					URL:        "http://example.com/" + title,
					Title:      title,
					UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
				},
			},
			expectErr: false,
		},
		{
			name: "create existing user website",
			web: model.UserWebsite{
				WebsiteUUID: uuid,
				UserUUID:    "def",
			},
			expect: model.UserWebsite{
				WebsiteUUID: uuid,
				UserUUID:    "def",
				GroupName:   title,
				AccessTime:  time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
				Website: model.Website{
					UUID:       uuid,
					URL:        "http://example.com/" + title,
					Title:      title,
					UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
				},
			},
			expectErr: false,
		},
		{
			name: "create new user website link to not exist website",
			web: model.UserWebsite{
				WebsiteUUID: "not-exist-uuid",
				UserUUID:    "new",
				GroupName:   "title",
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: model.UserWebsite{
				WebsiteUUID: "not-exist-uuid",
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
			t.Parallel()

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
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "update-user-website-uuid"
	title := "update user website"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
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
			web: model.UserWebsite{
				WebsiteUUID: uuid,
				UserUUID:    "def",
				GroupName:   title,
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expect: &model.UserWebsite{
				WebsiteUUID: uuid,
				UserUUID:    "def",
				GroupName:   title,
				AccessTime:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Website: model.Website{
					UUID:       uuid,
					URL:        "http://example.com/" + title,
					Title:      title,
					RawContent: "content",
					UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
				},
			},
			expectErr: false,
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
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "delete-user-website-uuid"
	title := "delete user website"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
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
			webUUID:   uuid,
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
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "find-user-websites-uuid"
	title := "find user websites"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
		db.Close()
	})

	tests := []struct {
		name      string
		userUUID  string
		expect    model.UserWebsite
		expectErr bool
	}{
		{
			name:     "find web of existing user",
			userUUID: "def",
			expect: model.UserWebsite{
				UserUUID:    "def",
				WebsiteUUID: uuid,
				GroupName:   title,
				AccessTime:  time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
				Website: model.Website{
					UUID:       uuid,
					URL:        "http://example.com/" + title,
					Title:      title,
					UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
				},
			},
			expectErr: false,
		},
		// {
		// 	name:      "find web of not existing user",
		// 	userUUID:  "not exist",
		// 	expect:    nil,
		// 	expectErr: false,
		// },
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := r.FindUserWebsites(test.userUUID)

			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}
			for _, userWeb := range result {
				if cmp.Equal(userWeb, test.expect) {
					return
				}
			}
			t.Errorf("result different from expected")
			t.Error(result)
			t.Error(test.expect)
		})
	}
}

func TestPsqlRepo_FindUserWebsitesByGroup(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "find-user-websites-group-uuid"
	title := "find user websites group"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
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
			group:    title,
			expect: model.WebsiteGroup{
				{
					UserUUID:    "def",
					WebsiteUUID: uuid,
					GroupName:   title,
					AccessTime:  time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
					Website: model.Website{
						UUID:       uuid,
						URL:        "http://example.com/" + title,
						Title:      title,
						UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
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
			group:     title,
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
	t.Parallel()

	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatalf("open database fail: %v", err)
	}

	r := NewRepo(db, &config.Config{})

	uuid := "find-user-website-uuid"
	title := "find user website"
	populateData(db, uuid, title)
	t.Cleanup(func() {
		db.Exec("delete from websites where uuid=$1", uuid)
		db.Exec("delete from user_websites where website_uuid=$1", uuid)
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
			webUUID:  uuid,
			expect: &model.UserWebsite{
				UserUUID:    "def",
				WebsiteUUID: uuid,
				GroupName:   title,
				AccessTime:  time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
				Website: model.Website{
					UUID:       uuid,
					URL:        "http://example.com/" + title,
					Title:      title,
					UpdateTime: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
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
			webUUID:   uuid,
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
