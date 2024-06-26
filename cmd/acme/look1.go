// #include <u.h>
// #include <libc.h>
// #include <draw.h>
// #include <thread.h>
// #include <cursor.h>
// #include <mouse.h>
// #include <keyboard.h>
// #include <frame.h>
// #include <fcall.h>
// #include <regexp.h>
// #include <9pclient.h>
// #include <plumb.h>
// #include <libsec.h>
// #include "dat.h"
// #include "fns.h"

package main

import (
	"bufio"
	"fmt"
	"time"

	"plramos.win/9fans/cmd/acme/internal/adraw"
	"plramos.win/9fans/cmd/acme/internal/alog"
	"plramos.win/9fans/cmd/acme/internal/bufs"
	"plramos.win/9fans/cmd/acme/internal/runes"
	"plramos.win/9fans/cmd/acme/internal/ui"
	"plramos.win/9fans/cmd/acme/internal/wind"
	"plramos.win/9fans/plan9"
	"plramos.win/9fans/plan9/client"
	"plramos.win/9fans/plumb"
)

var plumbeditfid *client.Fid

var nuntitled int

func plumbthread() {
	/*
	 * Loop so that if plumber is restarted, acme need not be.
	 */
	for {
		/*
		 * Connect to plumber.
		 */
		// TODO(rsc): plumbunmount()
		var fid *client.Fid
		for {
			var err error
			fid, err = plumb.Open("edit", plan9.OREAD|plan9.OCEXEC)
			if err == nil {
				break
			}
			time.Sleep(2 * time.Second)
		}
		bigLock() // TODO still racy
		plumbeditfid = fid
		ui.Plumbsendfid, _ = plumb.Open("send", plan9.OWRITE|plan9.OCEXEC)
		bigUnlock()

		/*
		 * Relay messages.
		 */
		bedit := bufio.NewReader(fid)
		for {
			m := new(plumb.Message)
			err := m.Recv(bedit)
			if err != nil {
				break
			}
			cplumb <- m
		}

		/*
		 * Lost connection.
		 */
		bigLock() // TODO still racy
		fid = ui.Plumbsendfid
		ui.Plumbsendfid = nil
		big.Unlock()
		fid.Close()

		bigLock() // TODO still racy
		fid = plumbeditfid
		plumbeditfid = nil
		big.Unlock()
		fid.Close()
	}
}

func startplumbing() {
	go plumbthread()
}

func plumbgetc(a interface{}, n int) rune {
	r := a.([]rune)
	if n > len(r) {
		return 0
	}
	return r[n]
}

func plumblook(m *plumb.Message) {
	if len(m.Data) >= bufs.Len {
		alog.Printf("insanely long file name (%d bytes) in plumb message (%.32s...)\n", len(m.Data), m.Data)
		return
	}
	var e ui.Expand
	e.Q0 = 0
	e.Q1 = 0
	if len(m.Data) == 0 {
		return
	}
	e.Arg = nil
	e.Bname = string(m.Data)
	e.Name = []rune(e.Bname)
	e.Jump = true
	e.A0 = 0
	e.A1 = 0
	addr := m.LookupAttr("addr")
	if addr != "" {
		r := []rune(addr)
		e.A1 = len(r)
		e.Arg = r
		e.Agetc = plumbgetc
	}
	adraw.Display.Top()
	ui.Openfile(nil, &e)
}

func plumbshow(m *plumb.Message) {
	adraw.Display.Top()
	w := ui.Makenewwindow(nil)
	ui.Winmousebut(w)
	name := m.LookupAttr("filename")
	if name == "" {
		nuntitled++
		name = fmt.Sprintf("Untitled-%d", nuntitled)
	}
	if name[0] != '/' && m.Dir != "" {
		name = fmt.Sprintf("%s/%s", m.Dir, name)
	}
	var rb [256]rune
	_, nr, _ := runes.Convert([]byte(name), rb[:], true)
	rs := runes.CleanPath(rb[:nr])
	wind.Winsetname(w, rs)
	r := make([]rune, len(m.Data))
	_, nr, _ = runes.Convert(m.Data, r, true)
	wind.Textinsert(&w.Body, 0, r[:nr], true)
	w.Body.File.SetMod(false)
	w.Dirty = false
	wind.Winsettag(w)
	wind.Textscrdraw(&w.Body)
	wind.Textsetselect(&w.Tag, w.Tag.Len(), w.Tag.Len())
	ui.OnNewWindow(w)
}
