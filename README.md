# go-lfile
A cross-platform Go package that provides a wrapper for os.File implementing RW mutex-like locking.

Transparently makes use of underlying system calls (flock/fcntl on Unix, LockFileEx on Windows) to provide locks on files, allowing file operations to synchronize within a process and between different processes.

## Install

Add package:
```
go get github.com/JSBanya/go-lfile
```

Include:
```
include (
  // (...)
  "github.com/JSBanya/go-lfile"
  // (...)
)
```


## Usage

### Creating a LockableFile
The lfile.LockableFile struct is a composition of an \*os.File, and thus implements all \*os.File functions. Creating a new lfile.LockableFile is done by simply calling ```func New(file *os.File) *LockableFile```:

Example:
```go
f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
if err != nil {
  log.Fatal(err)
}

lf := lfile.New(f)
name := lf.Name() // Same as calling f.Name() as lfile.LockableFile is a composition of *os.File
```

### Locking a file
To lock a lfile.LockableFile, simply call ```func (lf *LockableFile) Lock() error```.

By default, locking a file is done in blocking mode. This is similar to a mutex in that execution will pause if the file is already locked elsewhere (i.e. by another process). Non-blocking mode can be enabled by calling ```func (lf *LockableFile) DisableBlocking()```. In blocking is disabled, any calls to Lock that would normally pause execution (i.e. file is already locked) will instead return the error ```lfile.LOCK_CONFLICT```. Blocking mode can be enabled again at any time by calling ```func (lf *LockableFile) EnableBlocking()```

Two types of locks are provided: shared (read) locks, and exclusive (write) locks. Any number of processes may hold a read lock (shared lock) on a file, but only one process may hold a write (exclusive lock). Read locks are provided by calling ```func (lf *LockableFile) RLock() error``` and write locks are provided by calling ```func (lf *LockableFile) RWLock() error```

Example (blocking mode):
```go
// Open file as 'f' (not shown)

lf := lfile.New(f) // Default in blocking mode
if err := lf.RWLock(); err != nil { // Write lock (exclusive)
  log.Fatalf("Unable to lock file: %s", err)
}

// File is now locked and thread should have exclusive write access
lf.WriteString("my data")
```

Example (non-blocking mode):
```go
// Open file as 'f' (not shown)

lf := lfile.New(f)
lf.DisableBlocking()

err := lf.RWLock() // Write lock (exclusive)
if err == lfile.LOCK_CONFLICT {
  // File is locked already
  // Do something else (e.g. sleep and try again)
} else if err != nil {
  log.Fatalf("Unable to lock file: %s", err)
}

// File is now locked and thread should have exclusive write access
lf.WriteString("my data")
```

Example (read lock):
```go
// Open file as 'f' (not shown)
// Open same file as 'f2' (not shown)

lf := lfile.New(f)
if err := lf.RLock(); err != nil { // Read lock (shared)
  log.Fatalf("Unable to lock file: %s", err)
}

// (...)

lf2 := lfile.New(f2)
if err := lf2.RLock(); err != nil { // OK because previous lock was also a read (shared) lock
  log.Fatalf("Unable to lock file: %s", err)
}

// (...)

if err = lf.UnlockAndClose(); err != nil {
  log.Fatal(err)
}
 
if err = lf2.UnlockAndClose(); err != nil {
  log.Fatal(err)
}

```

Note: By default on Unix systems, lfile.LockableFile will use 'flock' to lock a file. For most purposes, this is likely desired, but 'fcntl' may instead be used by calling ```func (lf *LockableFile) UseFCNTL()``` (and can be switched back to flock via ```func (lf *LockableFile) UseFLOCK()```; on windows, these function calls will have no effect). Note that 'fcntl' locks files on a process level and thus may not have expected behavior in a multithreaded environment. See https://gavv.github.io/articles/file-locks/ for more details on the difference between flock and fcntl.

### Unlocking a file
Unlocking a file is done by simply calling ```func (lfile *LockableFile) Unlock() error```. If the file is not locked or locked by a different process, this function call will have no effect.

Simultaneously unlocking and closing a file can by done in one function call: ```func (lf *LockableFile) UnlockAndClose() error```

Example:
```go
// Open file as 'f' (not shown)

lf := lfile.New(f)
if err := lf.RWLock(); err != nil {
  log.Fatalf("Unable to lock file: %s", err)
}

// Do stuff

if err := lf.Unlock(); err != nil {
  log.Fatalf("Unable to unlock file: %s", err)
}
```
