package websiteupdate

import (
	"sync"

	"github.com/htchan/WebHistory/internal/model"
)

type Params struct {
	Web           *model.Website `json:"web"`
	ExecutionLock *sync.Mutex
}
