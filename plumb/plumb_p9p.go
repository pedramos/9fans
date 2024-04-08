//go:build !plan9
// +build !plan9

package plumb

import (
	"plramos.win/9fans/plan9/client"
)

func mountPlumb() {
	fsys, fsysErr = client.MountService("plumb")
}
