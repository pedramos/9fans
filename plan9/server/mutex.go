package server

import "sync"

// rwMutex is similar to sync.RWMutex but doesn't block.
type rwMutex struct {
	mu      sync.Mutex
	readers int
}

func (mu *rwMutex) rlock(onFail func()) bool {
	mu.mu.Lock()
	defer mu.mu.Unlock()
	if mu.readers == -1 {
		onFail()
		return false
	}
	mu.readers++
	return true
}

func (mu *rwMutex) runlock() {
	mu.mu.Lock()
	defer mu.mu.Unlock()
	mu.readers--
}

// lock tries to acquire an exclusive lock and reports whether
// it has succeeded. It doesn't block.
func (mu *rwMutex) lock(onFail func()) bool {
	mu.mu.Lock()
	defer mu.mu.Unlock()
	if mu.readers != 0 {
		onFail()
		return false
	}
	mu.readers = -1
	return true
}

func (mu *rwMutex) unlock() {
	mu.mu.Lock()
	defer mu.mu.Unlock()
	mu.readers = 0
}
