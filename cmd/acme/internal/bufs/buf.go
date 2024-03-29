package bufs

import (
	"sync"

	"plramos.win/9fans/cmd/acme/internal/runes"
)

const Len = 32 * 1024
const RuneLen = Len / runes.RuneSize

var runesPool = sync.Pool{
	New: func() interface{} { return make([]rune, RuneLen) },
}

func AllocRunes() []rune {
	return runesPool.Get().([]rune)
}

func FreeRunes(buf []rune) {
	if cap(buf) != RuneLen {
		panic("FreeRunes: wrong size")
	}
	runesPool.Put(buf[:RuneLen])
}

var bytesPool = sync.Pool{
	New: func() interface{} { return make([]byte, Len) },
}

func AllocBytes() []byte {
	return bytesPool.Get().([]byte)
}

func FreeBytes(buf []byte) {
	if cap(buf) != Len {
		panic("FreeRunes: wrong size")
	}
	bytesPool.Put(buf[:Len])
}
