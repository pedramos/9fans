package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"path"
	"sort"
	"testing"

	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestServerOpenRead(t *testing.T) {
	fs0, err := NewStaticFsys(map[string]StaticFile{
		"foo": {
			Content: []byte("bar"),
		},
		"info": {
			Entries: map[string]StaticFile{
				"version": {
					Content: []byte("something new"),
				},
				"other": {
					Content: bytes.Repeat([]byte("a"), 1024*1024),
				},
			},
		},
	})
	qt.Assert(t, err, qt.IsNil)
	c0, c1 := net.Pipe()
	errc := make(chan error, 1)
	go func() {
		err := Serve(context.Background(), c0, fs0)
		t.Logf("Serve finished; error: %v", err)
		c0.Close()
		errc <- err
	}()
	c, err := client.NewConn(c1)
	qt.Assert(t, err, qt.IsNil)
	defer c.Close()
	fs1, err := c.Attach(nil, "rog", "")
	qt.Assert(t, err, qt.IsNil)

	// Try reading a file.
	f, err := fs1.Open("/foo", plan9.OREAD)
	qt.Assert(t, err, qt.IsNil)
	data, err := io.ReadAll(f)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, string(data), qt.Equals, "bar")
	err = f.Close()
	qt.Assert(t, err, qt.IsNil)

	f, err = fs1.Open("/info", plan9.OREAD)
	qt.Assert(t, err, qt.IsNil)
	entries, err := f.Dirreadall()
	qt.Assert(t, err, qt.IsNil)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	qt.Assert(t, entries, qt.DeepEquals, []*plan9.Dir{{
		Name: "other",
		Qid: plan9.Qid{
			Path: 4,
		},
		Uid:    "noone",
		Gid:    "noone",
		Mode:   0o444,
		Length: 1024 * 1024,
	}, {
		Name: "version",
		Qid: plan9.Qid{
			Path: 5,
		},
		Uid:    "noone",
		Gid:    "noone",
		Mode:   0o444,
		Length: uint64(len("something new")),
	}})
	err = f.Close()
	qt.Assert(t, err, qt.IsNil)

	err = fs1.Close()
	qt.Assert(t, err, qt.IsNil)
	c.Release()
	err = <-errc
	qt.Assert(t, err, qt.IsNil)
}

func TestWalkDeep(t *testing.T) {
	file := StaticFile{
		Content: []byte("something"),
	}
	n := plan9.MAXWELEM * 3
	for i := n - 1; i >= 0; i-- {
		file = StaticFile{
			Entries: map[string]StaticFile{
				fmt.Sprint("dir", i): file,
			},
		}
	}
	fs0, err := NewStaticFsys(file.Entries)
	qt.Assert(t, err, qt.IsNil)

	c0, c1 := net.Pipe()
	errc := make(chan error, 1)
	go func() {
		err := Serve(context.Background(), c0, fs0)
		t.Logf("Serve finished; error: %v", err)
		c0.Close()
		errc <- err
	}()
	// We're just using NewConn for its connection init logic;
	// we'll actually do all the message sending and receiving ourselves.
	c, err := client.NewConn(c1)
	qt.Assert(t, err, qt.IsNil)
	defer c.Close()

	fs1, err := c.Attach(nil, "rog", "")
	qt.Assert(t, err, qt.IsNil)

	for i := 0; i < n-1; i++ {
		p := ""
		for j := 0; j <= i; j++ {
			p = path.Join(p, fmt.Sprint("dir", j))
		}
		info, err := fs1.Stat(p)
		qt.Assert(t, err, qt.IsNil)
		qt.Assert(t, info, qt.CmpEquals(cmpopts.IgnoreFields(plan9.Dir{}, "Qid")), &plan9.Dir{
			Name: path.Base(p),
			Mode: 0o555 | plan9.DMDIR,
			Uid:  "noone",
			Gid:  "noone",
		})
	}
	err = fs1.Close()
	qt.Assert(t, err, qt.IsNil)
	c.Release()
	err = <-errc
	qt.Assert(t, err, qt.IsNil)
}
