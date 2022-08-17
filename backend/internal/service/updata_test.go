package service

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/htchan/ApiParser"
	"github.com/htchan/WebHistory/internal/model"
	"github.com/htchan/WebHistory/internal/repo"
)

func Test_isTimeUpdated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		web    model.Website
		time   string
		expect bool
	}{
		{
			name:   "updated time",
			web:    model.Website{UpdateTime: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
			time:   "Sat, 1 Jan 2000 00:00:01 GMT",
			expect: true,
		},
		{
			name:   "not updated time",
			web:    model.Website{UpdateTime: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
			time:   "Sat, 1 Jan 2000 00:00:00 GMT",
			expect: false,
		},
		{
			name:   "time in wrong format",
			web:    model.Website{UpdateTime: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
			time:   "Sat, 1 Abc 2000 00:00:00 GMT",
			expect: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := isTimeUpdated(&test.web, test.time)
			if result != test.expect {
				t.Errorf("got different result as expect")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

func Test_isContentUpdated(t *testing.T) {
	t.Parallel()

	refArray := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		refArray = append(refArray, strconv.Itoa(i))
	}

	tests := []struct {
		name   string
		web    model.Website
		dates  []string
		expect bool
	}{
		{
			name:  "different dates",
			web:   model.Website{RawContent: strings.Join(refArray, model.SEP)},
			dates: append([]string{"0"}, refArray...),
		},
		{
			name:  "exact same dates",
			web:   model.Website{RawContent: strings.Join(refArray, model.SEP)},
			dates: refArray,
		},
		{
			name:  "exact same date with length model.DateLength at the beginning",
			web:   model.Website{RawContent: strings.Join(refArray, model.SEP)},
			dates: append(refArray[:model.DateLength], "999"),
		},
		{
			name:  "shorter dates",
			web:   model.Website{RawContent: strings.Join(refArray, model.SEP)},
			dates: refArray[:1],
		},
		{
			name:  "empty dates",
			web:   model.Website{RawContent: strings.Join(refArray, model.SEP)},
			dates: nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := isContentUpdated(&test.web, test.dates)
			if result != test.expect {
				t.Errorf("got different result as expect")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

func Test_isTitleUpdated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		web    model.Website
		title  string
		expect bool
	}{
		{
			name:   "different title",
			web:    model.Website{Title: "title"},
			title:  "new title",
			expect: true,
		},
		{
			name:   "exact same title",
			web:    model.Website{Title: "title"},
			title:  "title",
			expect: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := isTitleUpdated(&test.web, test.title)
			if result != test.expect {
				t.Errorf("got different result as expect")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

func Test_pruneResponse(t *testing.T) {
	temp := model.SEP
	model.SEP = ","
	t.Cleanup(func() {
		model.SEP = temp
	})

	tests := []struct {
		name   string
		resp   http.Response
		expect string
	}{
		{
			name: "replace content script, style, path node to SEP",
			resp: http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"a<scriptijk>ijk</script>b<styleijk>ijk</style>c<pathijk>ijk</path>d",
			)))},
			expect: "a,b,c,d",
		},
		{
			name: "replace <.*> to model sep",
			resp: http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"a<a>b</a>c<pre>d</pre>e",
			)))},
			expect: "a,b,c,d,e",
		},
		{
			name: "keep <title> in content",
			resp: http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"a<title>title</title>b",
			)))},
			expect: "a<title>title</title>b",
		},
		{
			name: "remove sep in beginning",
			resp: http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"<beginning>a<a>b</a>c<pre>d</pre>e</beginning>",
			)))},
			expect: "a,b,c,d,e",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := pruneResponse(&test.resp)
			if result != test.expect {
				t.Errorf("got different result as expect")
				t.Error(result)
				t.Error(test.expect)
			}
		})
	}
}

type MockClient struct {
	get func(string) (*http.Response, error)
}

func (m MockClient) Get(url string) (*http.Response, error) {
	return m.get(url)
}

func Test_Update(t *testing.T) {
	ApiParser.SetDefault(ApiParser.FromDirectory("../../assets/api_parser"))

	tempSEP := model.SEP
	tempDateLength := model.DateLength
	model.SEP = ","
	model.DateLength = 2
	t.Cleanup(func() {
		model.SEP = tempSEP
		model.DateLength = tempDateLength
	})

	refArray := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		refArray = append(refArray, strconv.Itoa(i))
	}

	tests := []struct {
		name       string
		r          repo.Repostory
		web        model.Website
		mockResp   *http.Response
		mockErr    error
		expectRepo repo.Repostory
		expectWeb  model.Website
		expectErr  bool
	}{
		{
			name: "updated title of web already have title",
			r:    repo.NewInMemRepo([]model.Website{{UUID: "uuid", Title: "title"}}, nil, nil),
			web:  model.Website{UUID: "uuid", Title: "title"},
			mockResp: &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"<title>title2</title>",
			)))},
			mockErr:    nil,
			expectRepo: repo.NewInMemRepo([]model.Website{{UUID: "uuid", Title: "title"}}, nil, nil),
			expectWeb:  model.Website{UUID: "uuid", Title: "title"},
		},
		{
			name: "updated title of web not having title",
			r:    repo.NewInMemRepo([]model.Website{{UUID: "uuid"}}, nil, nil),
			web:  model.Website{UUID: "uuid"},
			mockResp: &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"<title>title</title>",
			)))},
			mockErr:    nil,
			expectRepo: repo.NewInMemRepo([]model.Website{{UUID: "uuid", Title: "title"}}, nil, nil),
			expectWeb:  model.Website{UUID: "uuid", Title: "title"},
		},
		{
			name: "updated content",
			r:    repo.NewInMemRepo([]model.Website{{UUID: "uuid", RawContent: "11-1-1,22-2-2"}}, nil, nil),
			web:  model.Website{UUID: "uuid", RawContent: "11-1-1,22-2-2"},
			mockResp: &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"2222-2-2<a>33-3-3<a>44-4-4<a>",
			)))},
			mockErr:    nil,
			expectRepo: repo.NewInMemRepo([]model.Website{{UUID: "uuid", RawContent: "2222-2-2,33-3-3"}}, nil, nil),
			expectWeb:  model.Website{UUID: "uuid", RawContent: "2222-2-2,33-3-3"},
		},
		{
			name: "not updated content",
			r:    repo.NewInMemRepo([]model.Website{{UUID: "uuid", RawContent: "11-1-1,22-2-2"}}, nil, nil),
			web:  model.Website{RawContent: "11-1-1,22-2-2"},
			mockResp: &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(
				"11-1-1<a>22-2-2<a>33-3-3<a>44-4-4<a>",
			)))},
			mockErr:    nil,
			expectRepo: repo.NewInMemRepo([]model.Website{{UUID: "uuid", RawContent: "11-1-1,22-2-2"}}, nil, nil),
			expectWeb:  model.Website{RawContent: "11-1-1,22-2-2"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			client = MockClient{get: func(url string) (*http.Response, error) {
				return test.mockResp, test.mockErr
			}}

			err := Update(test.r, &test.web)

			if (err != nil) != test.expectErr {
				t.Errorf("got error: %v; want error: %v", err, test.expectErr)
			}

			if !cmp.Equal(test.r, test.expectRepo) {
				t.Error("got different repo as expect")
				t.Error(test.r)
				t.Error(test.expectRepo)
			}

			if !cmp.Equal(test.web, test.expectWeb) {
				t.Error("got different web as expect")
				t.Error(test.web)
				t.Error(test.expectWeb)
			}
		})
	}
}
