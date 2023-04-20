package server

import (
	"context"
	"errors"

	"9fans.net/go/plan9"
)

var (
	errNotImplemented = errors.New("operation not implemented")
	errPerm           = errors.New("permission denied")
)

// ErrorFsys implements Fsys by returning an error for all operations
// except Close. It's useful for embedding inside Fsys implementations
// when not all operations are implemented.
//
// It reports 64 for QidBits.
type ErrorFsys[F Fid] struct{}

func (ErrorFsys[F]) Auth(ctx context.Context, uname, aname string) (F, error) {
	return *new(F), errNotImplemented
}

func (ErrorFsys[F]) Attach(ctx context.Context, auth *F, uname, aname string) (F, error) {
	return *new(F), errNotImplemented
}

func (ErrorFsys[F]) Stat(ctx context.Context, f F) (plan9.Dir, error) {
	return plan9.Dir{}, errNotImplemented
}

func (ErrorFsys[F]) Wstat(ctx context.Context, f F, dir plan9.Dir) error {
	return errNotImplemented
}

func (ErrorFsys[F]) Walk(ctx context.Context, f F, name string) (F, error) {
	return *new(F), errNotImplemented
}

func (ErrorFsys[F]) Open(ctx context.Context, f F, mode uint8) (F, uint32, error) {
	return *new(F), 0, errNotImplemented
}

func (ErrorFsys[F]) Readdir(ctx context.Context, f F, dir []plan9.Dir, entryIndex int) (int, error) {
	return 0, errNotImplemented
}

func (ErrorFsys[F]) ReadAt(ctx context.Context, f F, buf []byte, off int64) (int, error) {
	return 0, errNotImplemented
}

func (ErrorFsys[F]) WriteAt(ctx context.Context, f F, buf []byte, off int64) (int, error) {
	return 0, errNotImplemented
}

func (ErrorFsys[F]) Remove(ctx context.Context, f F) error {
	return errNotImplemented
}

func (ErrorFsys[F]) Close() error {
	return nil
}

type qfid plan9.Qid

func (f qfid) Qid() plan9.Qid {
	return plan9.Qid(f)
}

type testErrorFsys struct {
	ErrorFsys[qfid]
}

func (testErrorFsys) Clone(f qfid) qfid {
	return f
}

func (testErrorFsys) Clunk(f qfid) {
}

var _ Fsys[qfid] = testErrorFsys{}
