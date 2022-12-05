// #include <u.h>
// #include <libc.h>
// #include <draw.h>
// #include <memdraw.h>

package memdraw

import "pedrolorgaramos.win/go/9fans/draw"

func allocmemsubfont(name string, n int, height int, ascent int, info []draw.Fontchar, i *Image) *subfont {
	f := new(subfont)
	f.n = n
	f.height = uint8(height)
	f.ascent = int8(ascent)
	f.info = info
	f.bits = i
	f.name = name
	return f
}

func freememsubfont(f *subfont) {
	if f == nil {
		return
	}
	Free(f.bits)
}
