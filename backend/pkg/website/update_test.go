package website

import (
	"time"
	"strings"
	"testing"
	"encoding/json"
	"io"
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/htchan/ApiParser"
)

func Test_checkTimeUpdate(t *testing.T) {
	t.Parallel()
	w := Website{
		UpdateTime: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	t.Run("before UpdateTime", func (t *testing.T) {
		t.Parallel()
		fail := w._checkTimeUpdate("Mon, 31 Dec 1999 00:00:00 GMT")
		if fail {
			t.Errorf("return %v", fail)
		}
	})

	t.Run("after UpdateTime", func (t *testing.T) {
		t.Parallel()
		success := w._checkTimeUpdate("Mon, 2 Jan 2000 00:00:00 GMT")
		if !success {
			t.Errorf("return %v", success)
		}
	})

	t.Run("invalid time format", func (t *testing.T) {
		t.Parallel()
		fail := w._checkTimeUpdate("2000-01-02 00:00:00 GMT")
		if fail {
			t.Errorf("return %v", fail)
		}
	})
}

func TestisUpdate(t *testing.T) {
	t.Parallel()

	w := Website{
		content: strings.Join([]string{"1", "2", "3"}, SEP),
	}

	t.Run("unchanged update dates", func (t *testing.T) {
		t.Parallel()
		fail := w.isUpdated([]string{"1", "2", "3"})
		if fail {
			t.Errorf("return %v", fail)
		}
	})

	t.Run("different update dates", func (t *testing.T) {
		t.Parallel()
		success := w.isUpdated([]string{"4", "5", "6"})
		if !success {
			t.Errorf("return %v", success)
		}
	})

	t.Run("different update dates order", func (t *testing.T) {
		t.Parallel()
		fail := w.isUpdated([]string{"3", "1", "2"})
		if fail {
			t.Errorf("return %v", fail)
		}
	})

	t.Run("cropped unchanged update dates", func (t *testing.T) {
		t.Parallel()
		fail := w.isUpdated([]string{"1", "2"})
		if fail {
			t.Errorf("return %v", fail)
		}
	})

	t.Run("more update dates", func (t *testing.T) {
		t.Parallel()
		success := w.isUpdated([]string{"1", "2", "3", "4"})
		if !success {
			t.Errorf("return %v", success)
		}
	})
}

func Test_checkBodyUpdate(t *testing.T) {
	t.Parallel()

	ApiParser.SetDefault(
		ApiParser.NewFormatSet(
			"website.info",
			`"(?P<Date>\d)"`,
			`"title":"(?P<Title>.*?)"`,
		),
	)
	w := Website{
		Title: "title",
		content: strings.Join([]string{"1", "2", "3"}, SEP),
	}

	t.Run("everything same", func (t *testing.T) {
		testW := w
		t.Parallel()
		responseSource := map[string]interface{} {
			"title": "title",
			"date": []string{"1", "2", "3"},
		}

		responseBody, err := json.Marshal(responseSource)
		if err != nil {
			t.Errorf("return error: %v", err)
		}
		result := testW._checkBodyUpdate(string(responseBody))
		if result {
			t.Errorf("return %v", result)
		}
	})

	t.Run("title updated", func (t *testing.T) {
		testW := w
		t.Parallel()
		responseSource := map[string]interface{} {
			"title": "title2",
			"date": []string{"1", "2", "3"},
		}

		responseBody, err := json.Marshal(responseSource)
		if err != nil {
			t.Errorf("return error: %v", err)
		}
		result := testW._checkBodyUpdate(string(responseBody))
		if !result {
			t.Errorf("return %v", result)
		}
	})

	t.Run("date changed", func (t *testing.T) {
		testW := w
		t.Parallel()
		responseSource := map[string]interface{} {
			"title": "title",
			"date": []string{"1", "2", "4"},
		}

		responseBody, err := json.Marshal(responseSource)
		t.Log(string(responseBody))
		if err != nil {
			t.Errorf("return error: %v", err)
		}
		result := testW._checkBodyUpdate(string(responseBody))
		if !result {
			t.Errorf("return %v", result)
		}
	})
}

func Test_pruneResponse(t *testing.T) {
	t.Parallel()
	t.Run("remove unwanted content", func (t *testing.T) {
		t.Parallel()
		body := "<script>abc\n<style>sss</style></script>\n<path abc>def</path>"
		response := &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(bytes.NewReader([]byte(body))),
		}
		result := pruneResponse(response)
		if result != "" {
			t.Errorf(`return "%v"`, result)
		}
	})

	t.Run("parse title and other tag", func (t *testing.T) {
		t.Parallel()
		body := "<div><title>abc</title></div>dev<div>def<div>"
		response := &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(bytes.NewReader([]byte(body))),
		}
		result := pruneResponse(response)
		if result != "<title>abc</title>\ndev\ndef" {
			t.Errorf(`return "%v"`, result)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if req.URL.String() == "/success" {
				responseSource := map[string]interface{} {
					"title": "title2",
					"date": []string{"1", "2", "3", "4"},
				}
				json.NewEncoder(res).Encode(responseSource)
			} else if req.URL.String() == "/unchanged" {
				responseSource := map[string]interface{} {
					"title": "title",
					"date": []string{"1", "2", "3"},
				}
				json.NewEncoder(res).Encode(responseSource)
			} else {
				res.Write([]byte("unknown"))
			}
		}),
	)

	ApiParser.SetDefault(
		ApiParser.NewFormatSet(
			"website.info",
			`"(?P<Date>\d)"`,
			`"title":"(?P<Title>.*?)"`,
		),
	)
	w := Website{
		Title: "title",
		content: strings.Join([]string{"1", "2", "3"}, SEP),
		UpdateTime: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	t.Run("fail to fetch website", func (t *testing.T) {
		testW := w
		t.Parallel()
		testW.URL = strings.ReplaceAll(server.URL, "http", "https")
		testW.Update()
		if !testW.UpdateTime.Equal(w.UpdateTime) {
			t.Errorf("changed update time")
		}
	})

	t.Run("successfully update will update the update time", func (t *testing.T) {
		testW := w
		t.Parallel()
		testW.URL = server.URL + "/success"
		testW.Update()
		if testW.UpdateTime.Equal(w.UpdateTime) {
			t.Errorf("not changed update time: %v", testW.UpdateTime)
		}
	})
}