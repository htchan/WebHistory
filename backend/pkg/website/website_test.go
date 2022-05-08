package website

import (
	"io"
	"os"
	"time"
	"testing"
)

func TestNewWebsite(t *testing.T) {
	t.Parallel()
	w := NewWebsite("http://localhost", "12345")

	if w.URL != "http://localhost" || w.UserUUID != "12345" ||
		w.Title != "Unknown" || w.GroupName != "Unknown" ||
		!w.UpdateTime.Before(time.Now()) || !w.AccessTime.Before(time.Now()) {
		t.Errorf("got: %v", w)
	}
}

func TestWithDB(t *testing.T) {
	t.Parallel()
	source, err := os.Open("../../assets/template.db")
	dest, err := os.Create("./test.db")
	_, err = io.Copy(dest, source)
	source.Close()
	dest.Close()
	if err != nil {
		t.Fatalf("fail to copy: %v", err)
	}
	db, err := OpenDatabase("test.db")
	if err != nil {
		t.Errorf("cannot open database: %v", err)
	}

	t.Run("Create", func (t *testing.T) {
		t.Run("brand new website", func (t *testing.T) {
			t.Parallel()
			url := "http://localhost/create"
			w := NewWebsite(url, "12345")
			err := w.Create(db)
			if err != nil {
				t.Errorf("return %v", err)
			}
			w, err = FindUserWebsite(db, "12345", url)
			if err != nil || w.URL != url {
				t.Errorf("find user web return website: %v, err: %v", w, err)
			}
		})

		t.Run("existing website with new user", func (t *testing.T) {
			t.Parallel()
			url := "http://localhost/create"
			w := NewWebsite(url, "23456")
			err := w.Create(db)
			if err != nil {
				t.Errorf("return %v", err)
			}
			w, err = FindUserWebsite(db, "23456", url)
			if err != nil || w.URL != url {
				t.Errorf("find user web return website: %v, err: %v", w, err)
			}
		})

		t.Run("existing website with existing user", func (t *testing.T) {
			t.Parallel()
			url := "http://localhost/create"
			w := NewWebsite(url, "12345")
			err := w.Create(db)
			if err != nil {
				t.Errorf("first create: return %v", err)
			}
			err = w.Create(db)
			if err != nil {
				t.Errorf("second create: return %v", err)
			}
		})
	})

	t.Run("Save", func (t *testing.T) {
		t.Run("update content", func (t *testing.T) {
			t.Parallel()
			url := "http://localhost/save"
			w := NewWebsite(url, "12345")
			w.Create(db)
			w.Title = "title2"
			err := w.Save(db)
			if err != nil {
				t.Errorf("return %v", err)
			}
			w, err = FindUserWebsite(db, "12345", url)
			if err != nil || w.URL != url {
				t.Errorf("find user web return website: %v, err: %v", w, err)
			}
		})
	})

	t.Run("Delete", func (t *testing.T) {
		t.Run("delete existing user website", func (t *testing.T) {
			t.Parallel()
			url := "http://localhost/delete"
			w := NewWebsite(url, "12345")
			w.Create(db)
			err := w.Delete(db)
			if err != nil {
				t.Errorf("return %v", err)
			}
			w, err = FindUserWebsite(db, "12345", url)
			if err == nil {
				t.Errorf("find user web return website: %v, err: %v", w, err)
			}
		})

		t.Run("delete not existing user website", func (t *testing.T) {
			t.Parallel()
			url := "http://localhost/delete"
			w := NewWebsite(url, "not_exist")
			w.Create(db)
			err := w.Delete(db)
			if err != nil {
				t.Errorf("return %v", err)
			}
			w, err = FindUserWebsite(db, "not_exist", url)
			if err == nil {
				t.Errorf("find user web return website: %v, err: %v", w, err)
			}
		})
	})

	t.Run("FindAllWebsites", func (t *testing.T) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()
			url := "http://localhost/all_web"
			w := NewWebsite(url, "12345")
			w.Create(db)
			ws, err := FindAllWebsites(db)
			if err != nil || len(ws) == 0 {
				t.Errorf("find user web return website: %v, err: %v", ws, err)
			}
		})
	})
	
	t.Run("FindAllUserWebsites", func (t *testing.T) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()
			url := "http://localhost/all_user_web"
			w := NewWebsite(url, "12345")
			w.Create(db)
			ws, err := FindAllUserWebsites(db, "12345")
			if err != nil || len(ws) == 0 {
				t.Errorf("find user web return website: %v, err: %v", ws, err)
			}
		})
	})

	t.Run("FindUserWebsite", func (t *testing.T) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()
			url := "http://localhost/all_user_web"
			w := NewWebsite(url, "12345")
			w.Create(db)
			ws, err := FindAllUserWebsites(db, "12345")
			if err != nil || len(ws) == 0 {
				t.Errorf("find user web return website: %v, err: %v", ws, err)
			}
		})
	})
	
	os.Remove("./test.db")
}

func TestWebsite(t *testing.T) {
	w := NewWebsite("http://a.b.localhost", "12345")
	w.Title = "title"
	w.GroupName = "group name"
	t.Run("Map", func (t *testing.T) {
		t.Parallel()
		result := w.Map()
		if result["url"] != "http://a.b.localhost" || result["title"] != "title" ||
			result["groupName"] != "group name" {
				t.Errorf("wrong map: %v", result)
			}
	})

	t.Run("Host", func (t *testing.T) {
		t.Parallel()
		result := w.Host()
		if result != "b.localhost" {
			t.Errorf("wrong host: %v", result)
		}
	})
}