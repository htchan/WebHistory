package websiteupdate

import (
	"testing"
	"time"

	"github.com/htchan/WebHistory/internal/repository"
	"github.com/stretchr/testify/assert"
)

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
