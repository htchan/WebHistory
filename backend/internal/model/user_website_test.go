package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewUserWebsite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		web                Website
		userUUID           string
		expectedGroupName  string
		expectedAccessTime time.Time
		want               UserWebsite
	}{
		{
			name:     "happy flow",
			web:      Website{UUID: "uuid"},
			userUUID: "user uuid",
			want: UserWebsite{
				WebsiteUUID: "uuid",
				UserUUID:    "user uuid",
				GroupName:   "",
				AccessTime:  time.Now().UTC().Truncate(time.Second),
				Website:     Website{UUID: "uuid"},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			userWeb := NewUserWebsite(test.web, test.userUUID)
			assert.Equal(t, test.want, userWeb)
		})
	}
}

func TestUserWebsite_MarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		web       UserWebsite
		want      string
		wantError error
	}{
		{
			name: "happy flow",
			web: UserWebsite{
				Website: Website{
					UUID:       "uuid",
					URL:        "http://example.com",
					Title:      "title",
					UpdateTime: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				UserUUID:   "user uuid",
				GroupName:  "group",
				AccessTime: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			want:      `{"uuid":"","user_uuid":"user uuid","url":"http://example.com","title":"title","group_name":"group","update_time":"2020-01-02T00:00:00 UTC","access_time":"2020-01-02T00:00:00 UTC"}`,
			wantError: nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := json.Marshal(test.web)
			assert.ErrorIs(t, err, test.wantError)
			assert.Equal(t, test.want, string(result))
		})
	}
}

func TestUserWebsites_WebsiteGroups(t *testing.T) {
	tests := []struct {
		name string
		webs UserWebsites
		want WebsiteGroups
	}{
		{
			name: "happy flow",
			webs: UserWebsites{
				UserWebsite{WebsiteUUID: "1", GroupName: "1"},
				UserWebsite{WebsiteUUID: "2", GroupName: "1"},
				UserWebsite{WebsiteUUID: "3", GroupName: "2"},
			},
			want: WebsiteGroups{
				WebsiteGroup{
					UserWebsite{WebsiteUUID: "1", GroupName: "1"},
					UserWebsite{WebsiteUUID: "2", GroupName: "1"},
				},
				WebsiteGroup{
					UserWebsite{WebsiteUUID: "3", GroupName: "2"},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			groups := test.webs.WebsiteGroups()
			assert.Equal(t, test.want, groups)
		})
	}
}
