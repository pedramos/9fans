package server

import (
	"context"

	"9fans.net/go/plan9"
)

// Fid represents a handle to a server-side file.
//
// The zero instance of a Fid is considered to represent
// "no fid" and should not otherwise be used as a valid value.
type Fid interface {
	comparable
	Qid() plan9.Qid
}

// Fsys represents the interface that must be implemented
// in order to provide a 9p server.
//
// Some methods (specifically Walk and Open) can choose
// whether or not to return a new instance of F. If they do return
// a new instance (determined by checking equality),
// the old one will be discarded by calling Clunk.
type Fsys[F Fid] interface {
	// Clone creates a copy of F. Note that this method will
	// be called more often than actual Tclone calls
	// (for example, a Twalk call will always invoke Clone
	// before walking).
	//
	// When an instance of F returned by Clone is no longer
	// in use, it will be discarded by calling Clunk.
	//
	// A fid that's been opened will never be cloned.
	Clone(f F) F

	// Clunk discards an instance of F. Clunk will never be called while there are any running
	// I/O methods on f.
	Clunk(f F)

	// Auth returns a new auth fid associated with the given user and attach name.
	// TODO should the returned fid be considered open?
	Auth(ctx context.Context, uname, aname string) (F, error)

	// Attach attaches to a new tree, returning a new
	// instance of F representing the root.
	// If auth is non-nil, it holds the auth fid.
	Attach(ctx context.Context, auth *F, uname, aname string) (F, error)

	Stat(ctx context.Context, f F) (plan9.Dir, error)
	Wstat(ctx context.Context, f F, dir plan9.Dir) error

	// Walk walks to the named element within the directory
	// represented by f and returns a handle to that element.
	Walk(ctx context.Context, f F, name string) (F, error)

	// Open prepares a fid for I/O.
	// After it's been opened, no methods will be called other
	// than Readdir (if it's a directory), ReadAt or WriteAt (if it's a file)
	// or Clunk.
	// Open returns the opened file and the its associated iounit.
	Open(ctx context.Context, f F, mode uint8) (F, uint32, error)

	// Readdir reads directory entries from an open directory into
	// dir, starting at the number of entries into the directory.
	// It returns the number of entries read.
	Readdir(ctx context.Context, f F, dir []plan9.Dir, entryIndex int) (int, error)

	ReadAt(ctx context.Context, f F, buf []byte, off int64) (int, error)
	WriteAt(ctx context.Context, f F, buf []byte, off int64) (int, error)

	// Remove removes the file represented by f. Unlike 9p remove,
	// this does not imply a clunk - the Clunk method will be explicitly
	// called immediately after Remove.
	Remove(ctx context.Context, f F) error

	// Close is called when the entire server instance is being torn down.
	Close() error
}

// Synchronous returns the set of message types for which
// operations on f  will always return immediately.
// Operations in this list will be called synchronously
// within the server (no other methods will be called until the method
// returns)
//Synchronous(f F) OpSet

// QidBits returns how many bits of Qid path space
// this server uses (counting from least significant upwards).
// This enables Fsys implementations to wrap other Fsys
// without worrying about QID path clashes.
//QidBits() int
