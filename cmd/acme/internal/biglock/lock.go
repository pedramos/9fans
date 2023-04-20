package biglock

import (
	"9fans.net/go/cmd/acme/internal/sync"
)

var big sync.Mutex

func init() {
	big.SetName("bigLock")
}
