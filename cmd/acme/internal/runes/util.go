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

package runes

import (
	"io"
	"unicode/utf8"
)

type Text interface {
	Len() int
	RuneAt(pos int) rune
	Slice(q0, q1 int) []rune
}

type Runes []rune

func (r Runes) Len() int                { return len(r) }
func (r Runes) RuneAt(pos int) rune     { return r[pos] }
func (r Runes) Slice(q0, q1 int) []rune { return r[q0:q1] }

type RuneReader struct {
	start, curr, end int
	t                Text
}

func NewRuneReader(t Text, q0, q1 int) *RuneReader {
	return &RuneReader{t: t, start: q0, curr: q0, end: q1}
}

func (rr *RuneReader) Seek(q0 int) {
	rr.curr = q0
	rr.start = q0
}

func (rr *RuneReader) ReadRune() (r rune, size int, err error) {
	r = rr.t.RuneAt(rr.curr)

	rr.curr++
	if rr.curr >= rr.end {
		err = io.EOF
	}
	return r, utf8.RuneLen(r), err
}

func (rr *RuneReader) Start() int { return rr.start }
