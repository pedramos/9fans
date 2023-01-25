package main

import (
	"plramos.win/9fans/cmd/acme/internal/wind"
	"plramos.win/9fans/cmd/internal/base"
	"plramos.win/9fans/draw"
	"plramos.win/9fans/plan9"
	"plramos.win/9fans/plumb"
)

const (
	Qdir = iota
	Qacme
	Qcons
	Qconsctl
	Qdraw
	Qeditout
	Qindex
	Qlabel
	Qlog
	Qnew
	QWaddr
	QWbody
	QWctl
	QWdata
	QWeditout
	QWerrors
	QWevent
	QWrdsel
	QWwrsel
	QWtag
	QWxdata
	QMAX
)

type Timer struct {
	dt     int
	cancel int
	c      chan int
	next   *Timer
}

type Dirtab struct {
	name string
	typ  uint8
	qid  int
	perm int
}

type Fid struct {
	fid    int
	busy   bool
	open   bool
	qid    plan9.Qid
	w      *wind.Window
	dir    []Dirtab
	next   *Fid
	mntdir *base.Mntdir
	rpart  []byte
	logoff int64
}

type Xfid struct {
	arg   interface{}
	fcall *plan9.Fcall
	next  *Xfid
	c     chan func(*Xfid)
	f     *Fid
	// buf     *uint8
	flushed bool
}

var screen *draw.Image
var keyboardctl *draw.Keyboardctl
var timerpid int
var fsyspid int
var cputype string
var home string
var dodollarsigns bool

type Waitmsg struct {
	pid int
	msg string
}

var (
	cplumb     = make(chan *plumb.Message)
	cxfidalloc = make(chan *Xfid) // TODO bidi
	cxfidfree  = make(chan *Xfid)
	cnewwindow = make(chan *wind.Window) // TODO bidi
	mouseexit0 chan int
	mouseexit1 chan int
	cerr       = make(chan []byte)
	cwarn      = make(chan int, 1)
)

// #define	STACK	65536
