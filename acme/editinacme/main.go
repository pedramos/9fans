// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Editinacme can be used as $EDITOR in a Unix environment.
//
// Usage:
//
//	editinacme <file1> [<file2>...]
//
// Editinacme uses the plumber to ask acme to open the file,
// waits until the file's acme window is deleted, and exits.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"plramos.win/9fans/acme"
)

var openall = flag.Bool("a", false, "Open all files at once, otherwise it will open one at a time")

func main() {
	log.SetFlags(0)
	log.SetPrefix("editinacme: ")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: editinacme file1 [file2]...\n")
		os.Exit(2)
	}
	flag.Parse()

	r, err := acme.Log()
	if err != nil {
		log.Fatal(err)
	}
	var files = make(map[string]struct{})
	for _, file := range flag.Args() {
		fullpath, err := filepath.Abs(file)
		if err != nil {
			log.Fatal(err)
		}
		file = fullpath

		log.Printf("editing %s", file)

		out, err := exec.Command("plumb", "-d", "edit", file).CombinedOutput()
		if err != nil {
			log.Fatalf("executing plumb: %v\n%s", err, out)
		}
		if !*openall {
			for {
				ev, err := r.Read()
				if err != nil {
					log.Fatalf("reading acme log: %v", err)
				}
				if ev.Op == "del" && ev.Name == file {
					break
				}
			}
		} else {
			files[file] = struct{}{}
		}
	}

	for len(files) > 0 {
		ev, err := r.Read()
		if err != nil {
			log.Fatalf("reading acme log: %v", err)
		}
		if _, found := files[ev.Name]; ev.Op == "del" && found {
			delete(files, ev.Name)
		}
	}
}
