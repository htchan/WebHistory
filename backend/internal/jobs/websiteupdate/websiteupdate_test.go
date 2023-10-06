package websiteupdate

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/htchan/WebHistory/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	leak := flag.Bool("leak", false, "check for memory leaks")
	flag.Parse()

	if *leak {
		goleak.VerifyTestMain(m)
	} else {
		os.Exit(m.Run())
	}
}

func TestSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rpo           repository.Repostory
		sleepInterval time.Duration
		wantJob       *Job
	}{
		{
			name:          "happy flow",
			rpo:           nil,
			sleepInterval: 5 * time.Second,
			wantJob:       &Job{rpo: nil, sleepInterval: 5 * time.Second},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			scheduler := Setup(test.rpo, test.sleepInterval)
			assert.Equal(t, test.wantJob, scheduler.job)
		})
	}
}
