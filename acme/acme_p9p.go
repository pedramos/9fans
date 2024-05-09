//go:build !plan9
// +build !plan9

package acme

import "plramos.win/9fans/plan9/client"

func mountAcme() {
	fs, err := client.MountService("acme")
	fsys = fs
	fsysErr = err
}
