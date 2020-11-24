// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris aix

package lfile

import (
	"io"
	"log"
	"syscall"
)

func (lf *LockableFile) UseFCNTL() {
	lf.unixLockType = FCNTL
}

func (lf *LockableFile) UseFLOCK() {
	lf.unixLockType = FLOCK
}

func (lf *LockableFile) lock(exclusive bool) error {
	fd := lf.Fd()

	var err error
	switch lf.unixLockType {
	case FCNTL:
		{
			cmd := syscall.F_SETLK
			if lf.blocking {
				cmd = syscall.F_SETLKW
			}

			var l_type int16 = syscall.F_RDLCK
			if exclusive {
				l_type = syscall.F_WRLCK
			}

			flock := &syscall.Flock_t{
				Type:   l_type,
				Whence: io.SeekStart,
				Start:  0,
				Len:    0,
			}

			err = syscall.FcntlFlock(fd, cmd, flock)
			if err != nil {
				if errno, ok := err.(syscall.Errno); ok && (errno == syscall.EACCES || errno == syscall.EAGAIN) {
					return LOCK_CONFLICT
				}

				// Unhandled error
				return err
			}
		}
	case FLOCK:
		{
			operation := syscall.LOCK_SH
			if exclusive {
				operation = syscall.LOCK_EX
			}

			if !lf.blocking {
				operation |= syscall.LOCK_NB
			}

			err = syscall.Flock(int(fd), operation)
			if err != nil {
				if errno, ok := err.(syscall.Errno); ok && errno == syscall.EWOULDBLOCK {
					return LOCK_CONFLICT
				}

				// Unhandled error
				return err
			}
		}
	default:
		log.Fatal("Unrecognized lock type.")
	}

	return nil
}

func (lf *LockableFile) RLock() error {
	return lf.lock(false)
}

func (lf *LockableFile) RWLock() error {
	return lf.lock(true)
}

func (lfile *LockableFile) Unlock() error {
	fd := lfile.Fd()

	var err error
	switch lfile.unixLockType {
	case FCNTL:
		{
			flock := &syscall.Flock_t{
				Type:   syscall.F_UNLCK,
				Start:  0,
				Len:    0,
				Whence: 0, // SEEK_SET
			}

			err = syscall.FcntlFlock(fd, syscall.F_SETLKW, flock)
			if err != nil {
				return err
			}
		}
	case FLOCK:
		{
			err = syscall.Flock(int(fd), syscall.LOCK_UN)
			if err != nil {
				return err
			}
		}
	default:
		log.Fatal("Unrecognized lock type.")
	}

	return nil
}
