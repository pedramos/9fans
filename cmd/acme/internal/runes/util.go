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

type Text interface {
	Len() int
	RuneAt(pos int) rune
	Slice(q0, q1 int) []rune
}

type Runes []rune

func (r Runes) Len() int                { return len(r) }
func (r Runes) RuneAt(pos int) rune     { return r[pos] }
func (r Runes) Slice(q0, q1 int) []rune { return r[q0:q1] }
