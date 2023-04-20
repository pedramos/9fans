package server

import (
	"context"
	"fmt"
	stdpath "path"
	"sort"
	"strings"

	"9fans.net/go/plan9"
)

var errNotFound = fmt.Errorf("file not found")

type staticFsys struct {
	ErrorFsys[*staticFileQ]
	root *staticFileQ
}

type StaticFile struct {
	// Entries holds the set of entries in a directory.
	// If it's nil, the StaticFile represents a regular
	// file and Content holds its contents.
	Entries    map[string]StaticFile
	Content    []byte
	Executable bool
}

type staticFileQ struct {
	qid        plan9.Qid
	name       string
	perm       uint
	executable bool
	content    []byte
	entries    []*staticFileQ
}

type StaticFsys = Fsys[*staticFileQ]

// NewStaticFsys returns an instance of StaticFsys that serves
// static read-only content. The given entries hold the contents of the root
// directory.
func NewStaticFsys(entries map[string]StaticFile) (StaticFsys, error) {
	root := StaticFile{
		Entries: entries,
	}
	rootq, _, err := calcQids(".", root, "", 1)
	if err != nil {
		return nil, fmt.Errorf("bad file tree: %v", err)
	}
	return &staticFsys{
		root: rootq,
	}, nil
}

func (f *staticFileQ) Qid() plan9.Qid {
	return f.qid
}

func (fs *staticFsys) Clone(f *staticFileQ) *staticFileQ {
	return f
}

func (fs *staticFsys) Clunk(f *staticFileQ) {
}

func (fs *staticFsys) Attach(ctx context.Context, _ **staticFileQ, uname, aname string) (*staticFileQ, error) {
	return fs.root, nil
}

func (fs *staticFsys) Stat(ctx context.Context, f *staticFileQ) (plan9.Dir, error) {
	return fs.makeDir(f), nil
}

func (fs *staticFsys) makeDir(f *staticFileQ) plan9.Dir {
	m := plan9.Perm(0o444)
	if f.executable || f.entries != nil {
		m |= 0o111
	}
	length := uint64(0)
	if f.entries != nil {
		m |= plan9.DMDIR
	} else {
		length = uint64(len(f.content))
	}
	return plan9.Dir{
		Qid:    f.qid,
		Name:   f.name,
		Mode:   m,
		Length: length,
		Uid:    "noone",
		Gid:    "noone",
	}
}

func (fs *staticFsys) Walk(ctx context.Context, f *staticFileQ, name string) (*staticFileQ, error) {
	for _, e := range f.entries {
		if e.name == name {
			return e, nil
		}
	}
	return nil, errNotFound
}

func (fs *staticFsys) Readdir(ctx context.Context, f *staticFileQ, dir []plan9.Dir, index int) (int, error) {
	if index >= len(f.entries) {
		index = len(f.entries)
	}
	for i, e := range f.entries[index:] {
		dir[i] = fs.makeDir(e)
	}
	return len(f.entries) - index, nil
}

func (fs *staticFsys) Open(ctx context.Context, f *staticFileQ, mode uint8) (*staticFileQ, uint32, error) {
	// caller has already checked file perms, so just allow it.
	return f, 8192, nil
}

func (fs *staticFsys) ReadAt(ctx context.Context, f *staticFileQ, buf []byte, off int64) (int, error) {
	if off > int64(len(f.content)) {
		off = int64(len(f.content))
	}
	return copy(buf, f.content[off:]), nil
}

func validName(s string) bool {
	return !strings.ContainsAny(s, "/")
}

func calcQids(fname string, f StaticFile, path string, qpath uint64) (_ *staticFileQ, maxQpath uint64, err error) {
	if !validName(fname) {
		return nil, 0, fmt.Errorf("file name %q in directory %q isn't valid", fname, path)
	}
	path = stdpath.Join(path, fname)
	if f.Content != nil && f.Entries != nil {
		return nil, 0, fmt.Errorf("%q has both content and entries set", path)
	}
	qtype := uint8(0)
	if f.Entries != nil {
		qtype = plan9.QTDIR
	}
	qf := &staticFileQ{
		qid: plan9.Qid{
			Path: qpath,
			Type: qtype,
		},
		name:       fname,
		executable: f.Executable,
		content:    f.Content,
	}
	qpath++
	if f.Entries == nil {
		return qf, qpath, nil
	}
	// sort by name for predictability of tests.
	names := make([]string, 0, len(f.Entries))
	for name := range f.Entries {
		names = append(names, name)
	}
	sort.Strings(names)
	qf.entries = make([]*staticFileQ, len(names))
	for i, name := range names {
		entry := f.Entries[name]
		e, qp, err := calcQids(name, entry, path, qpath)
		if err != nil {
			return nil, 0, err
		}
		qf.entries[i] = e
		qpath = qp
	}
	return qf, qpath, nil
}
