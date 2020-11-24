package lfile

import (
	"errors"
	"os"
)

var (
	LOCK_CONFLICT = errors.New("File already locked")
)

type LockType int

const (
	FLOCK LockType = 0 // Thread-safe if applied to different open() calls; May not allow NFS
	FCNTL LockType = 1 // Not thread-safe; Allows NFS; Max compatibility
)

type LockableFile struct {
	*os.File // Composition; effectively acts as a standard os.File

	blocking     bool
	unixLockType LockType
}

// Converts an existing *os.File to a *LockableFile
func New(file *os.File) *LockableFile {
	lfile := &LockableFile{
		File:     file,
		blocking: true,
	}
	lfile.UseFLOCK()
	return lfile
}

func (lf *LockableFile) EnableBlocking() {
	lf.blocking = true
}

func (lf *LockableFile) DisableBlocking() {
	lf.blocking = false
}

// Unlocks any existing lock and calls file close
func (lf *LockableFile) UnlockAndClose() error {
	if err := lf.Unlock(); err != nil {
		return err
	}

	return lf.Close()
}
