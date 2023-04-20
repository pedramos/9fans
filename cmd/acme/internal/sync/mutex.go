package sync

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Pool = sync.Pool

type Cond = sync.Cond

type Mutex struct {
	initOnce sync.Once
	c        chan struct{}
	goid     int64
	name     string
	callers  [6]uintptr
	ncallers int
}

func (m *Mutex) init() {
	m.initOnce.Do(func() {
		m.c = make(chan struct{}, 1)
	})
}

func (mu *Mutex) SetName(s string) {
	mu.Lock()
	defer mu.Unlock()
	mu.name = s
}

func (mu *Mutex) Lock() {
	mu.init()
	for {
		timer := time.NewTimer(20 * time.Second)
		defer timer.Stop()
		select {
		case mu.c <- struct{}{}:
			mu.goid = runtime.GoroutineID()
			mu.ncallers = runtime.Callers(2, mu.callers[:])
			return
		case <-timer.C:
			var buf strings.Builder
			fmt.Fprintf(&buf, "goroutine %d\n", runtime.GoroutineID())
			fmt.Fprintf(&buf, "\tprobable deadlock acquiring mutex %q\n", mu.name)
			fmt.Fprintf(&buf, "\toriginally acquired by goroutine %d\n", mu.goid)
			fmt.Fprintf(&buf, "\tcallers:\n")

			frames := runtime.CallersFrames(mu.callers[:mu.ncallers])
			for {
				fr, more := frames.Next()
				fmt.Fprintf(&buf, "\t\t%s:%d: %s\n", fr.File, fr.Line, fr.Function)
				if !more {
					break
				}
			}
			log.Printf("%s", buf.String())
		}
	}
}

func (mu *Mutex) Unlock() {
	mu.goid = 0
	mu.ncallers = 0
	<-mu.c
}
