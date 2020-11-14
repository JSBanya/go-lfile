// +build windows

package lfile

import (
	"syscall"
	"unsafe"
	"errors"
)

var (
	K32 = syscall.NewLazyDLL("kernel32.dll")
	LockFileEx = K32.NewProc("LockFileEx").Addr()
	UnlockFileEx = K32.NewProc("UnlockFileEx").Addr()
)

const (
	LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
	LOCKFILE_EXCLUSIVE_LOCK = 0x00000002
	MAXDWORD = ^uintptr(0)
)

func (lf *LockableFile) UseFCNTL() {
	// Not supported
}

func (lf *LockableFile) UseFLOCK() {
	// Not supported
}

func (lf *LockableFile) Lock() error {
	fd := lf.Fd()

	dwFlags := LOCKFILE_EXCLUSIVE_LOCK
	if !lf.blocking {
		dwFlags |= LOCKFILE_FAIL_IMMEDIATELY
	}

	var ol syscall.Overlapped
	r1, _, errno := syscall.Syscall6(LockFileEx, 6, fd, dwFlags, 0, 0, MAXDWORD, uintptr(unsafe.Pointer(ol)))
	if r1 == 0 { // "If the function succeeds, the return value is nonzero (TRUE)."
		if errno == syscall.ERROR_IO_PENDING {
			return LOCK_CONFLICT
		} else {
			return errors.New(errno.Error())
		}
	}
	
	return nil
}

func (lf *LockableFile) Unlock() error {
	fd := lf.Fd()

	var ol syscall.Overlapped
	r1, _, errno := syscall.Syscall6(UnlockFileEx, 5, fd, 0, 0, MAXDWORD, uintptr(unsafe.Pointer(ol)), 0)
	if r1 == 0 { // If the function succeeds, the return value is nonzero.
		return errors.New(errno.Error())
	}
	
	return nil
}