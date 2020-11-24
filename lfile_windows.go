// +build windows

package lfile

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	K32          = syscall.NewLazyDLL("kernel32.dll")
	LockFileEx   = K32.NewProc("LockFileEx")
	UnlockFileEx = K32.NewProc("UnlockFileEx")
)

const (
	LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
	LOCKFILE_EXCLUSIVE_LOCK   = 0x00000002
	LOCK_CONFLICT_ERRNO = syscall.Errno(33)
	ALREADY_UNLOCKED_ERRNO = syscall.Errno(158)
)

func (lf *LockableFile) UseFCNTL() {
	// Not supported
}

func (lf *LockableFile) UseFLOCK() {
	// Not supported
}

func (lf *LockableFile) lock(exclusive bool) error {
	dwFlags := 0
	if exclusive {
		dwFlags |= LOCKFILE_EXCLUSIVE_LOCK
	}

	if !lf.blocking {
		dwFlags |= LOCKFILE_FAIL_IMMEDIATELY
	}

	var hFile uintptr = lf.Fd()
	var lpOverlapped syscall.Overlapped
	lpOverlappedPtr := uintptr(unsafe.Pointer(&lpOverlapped))

	r1, _, errno := LockFileEx.Call(hFile, uintptr(dwFlags), 0, 1, 0, lpOverlappedPtr)
	if r1 == 0 { // "If the function succeeds, the return value is nonzero (TRUE)."
		// According to docs, LockFileEx returns ERROR_IO_PENDING if the call would have
		// blocked but LOCKFILE_FAIL_IMMEDIATELY was specified
		// However, in practice, errno is typically the value 33 instead, which is equivalent but has no
		// named counterpart. Thus, we check both. 
		if errno == LOCK_CONFLICT_ERRNO || errno == syscall.ERROR_IO_PENDING {
			return LOCK_CONFLICT
		} else {
			return errors.New(errno.Error())
		}
	}

	return nil
}

func (lf *LockableFile) RLock() error {
	return lf.lock(false)
}

func (lf *LockableFile) RWLock() error {
	return lf.lock(true)
}

func (lf *LockableFile) Unlock() error {
	var hFile uintptr = lf.Fd()
	var lpOverlapped syscall.Overlapped
	lpOverlappedPtr := uintptr(unsafe.Pointer(&lpOverlapped))

	r1, _, errno := UnlockFileEx.Call(hFile, 0, 1, 0, lpOverlappedPtr)
	if r1 == 0 && errno != ALREADY_UNLOCKED_ERRNO { // If the function succeeds, the return value is nonzero.
		return errors.New(errno.Error())
	}

	return nil
}
