// #include <u.h>
// #include <libc.h>
// #include <draw.h>
// #include <thread.h>
// #include <cursor.h>
// #include <mouse.h>
// #include <keyboard.h>
// #include <frame.h>
// #include <fcall.h>
// #include <plumb.h>
// #include <libsec.h>
// #include "dat.h"
// #include "fns.h"

package ui

import (
	"pedrolorgaramos.win/go/9fans/cmd/acme/internal/adraw"
	"pedrolorgaramos.win/go/9fans/cmd/acme/internal/file"
	"pedrolorgaramos.win/go/9fans/cmd/acme/internal/wind"
	"pedrolorgaramos.win/go/9fans/draw"
)

func WinresizeAndMouse(w *wind.Window, r draw.Rectangle, safe, keepextra bool) int {
	mouseintag := Mouse.Point.In(w.Tag.All)
	mouseinbody := Mouse.Point.In(w.Body.All)

	y := wind.Winresize(w, r, safe, keepextra)

	// If mouse is in tag, pull up as tag closes.
	if mouseintag && !Mouse.Point.In(w.Tag.All) {
		p := Mouse.Point
		p.Y = w.Tag.All.Max.Y - 3
		adraw.Display.MoveCursor(p)
	}

	// If mouse is in body, push down as tag expands.
	if mouseinbody && Mouse.Point.In(w.Tag.All) {
		p := Mouse.Point
		p.Y = w.Tag.All.Max.Y + 3
		adraw.Display.MoveCursor(p)
	}

	return y
}

func Wintype(w *wind.Window, t *wind.Text, r rune) {
	Texttype(t, r)
	if t.What == wind.Body {
		for i := 0; i < len(t.File.Text); i++ {
			wind.Textscrdraw(t.File.Text[i])
		}
	}
	wind.Winsettag(w)
}

var fff *file.File
