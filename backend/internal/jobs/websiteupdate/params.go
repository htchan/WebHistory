package websiteupdate

import (
	"sync"

	"github.com/htchan/WebHistory/internal/model"
)

type Params struct {
	web           *model.Website
	executionLock *sync.Mutex
}
