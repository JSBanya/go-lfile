package lfile

import (
	"testing"
	"io/ioutil"
	"os"
	"log"
	"time"
)

func createTemp() (string, error) {
	log.Println("Creating temporary file for testing.")
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	tempFileName := tempFile.Name()
	log.Printf("Temporary file created at %s", tempFileName)
	
	return tempFileName, nil
}

func TestNew(t *testing.T) {
	log.Println("TestNew running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	lf := New(f)
	if tempFileName != lf.Name() {
		t.Fatalf("expected %s, got %s", tempFileName, lf.Name())
	}
}

func TestUnlockOnNonLockedFileFLOCK(t *testing.T) {
	log.Println("TestUnlockOnNonLockedFileFLOCK running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	lf := New(f)
	if err = lf.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestUnlockOnFileLockedByOther(t *testing.T) {
	log.Println("TestUnlockOnFileLockedByOther running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	f2, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	lf := New(f)
	lf2 := New(f2)

	if err = lf.RWLock(); err != nil {
		t.Fatal(err)
	}

	if err = lf2.Unlock(); err != nil {
		t.Fatal(err)
	}

	if err = lf.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestUnlockOnNonLockedFileFCNTL(t *testing.T) {
	log.Println("TestUnlockOnNonLockedFileFCNTL running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	lf := New(f)
	lf.UseFCNTL()
	if err = lf.Unlock(); err != nil {
		t.Fatal(err)
	}
}

func TestClose(t *testing.T) {
	log.Println("TestClose running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}

	lf := New(f)
	if err = lf.RWLock(); err != nil {
		t.Fatal(err)
	}

	if err = lf.UnlockAndClose(); err != nil {
		t.Fatal(err)
	}

	if err = f.Close(); err == nil  {
		t.Fatal("File should have already been closed.")
	}
}

func TestNonblockingErrors(t *testing.T) {
	log.Println("TestNonblockingErrors running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	f2, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	lf := New(f)
	lf.DisableBlocking()

	lf2 := New(f2)
	lf2.DisableBlocking()

	if err = lf.RWLock(); err != nil {
		t.Fatal(err)
	}

	err = lf2.RWLock();
	if err == nil {
		t.Fatal("Expected LOCK_CONFLICT error")
	} else if err != LOCK_CONFLICT {
		t.Fatal(err)
	}

	if err = lf.Unlock(); err != nil {
		t.Fatal(err)
	}

	if err = lf2.RWLock(); err != nil {
		t.Fatal(err)
	}

	if err = lf2.Unlock(); err != nil {
		t.Fatal(err)
	}
}


// With high probability, tests whether mutex correctly locks file
// Only possible to test flock on unix systems as fcntl does not work with multithreading
func TestBlockingLockAndUnlock(t *testing.T) {
	log.Println("TestBlockingLockAndUnlock running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}

	fmux := New(f)
	fmux.UseFLOCK() // No-op on windows

	err = fmux.RWLock()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		// Thread 2
		f2, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			t.Fatal(err)
		}
	
		fmux2 := New(f2)
		fmux2.UseFLOCK()

		err = fmux2.RWLock()
		if err != nil {
			t.Fatal(err)
		}

		_, err = f2.WriteString("thread2")
		if err != nil {
			t.Fatal(err)
		}

		fmux2.Unlock()
		f2.Close()
	}()

	time.Sleep(2 * time.Second)
	_, err = f.WriteString("thread1 ")
	if err != nil {
		t.Fatal(err)
	}

	fmux.Unlock() // Should cause Thread 2 to unlock
	f.Close()

	time.Sleep(2 * time.Second) // Wait for Thread 2 to finish

	// Check contents of file
	f, err = os.Open(tempFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "thread1 thread2" {
		t.Fatalf("expected \"thread1 thread2\", got %s", string(content))
	}
}

func TestMultipleReadLocks(t *testing.T) {
	log.Println("TestMultipleReadLocks running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}	

	for i := 0; i < 4; i++ {
		f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			t.Fatal(err)
		}
	
		lf := New(f)
		if err = lf.RLock(); err != nil {
			t.Fatal(err)
		}
		defer lf.UnlockAndClose()
	}
}

func TestExclusive(t *testing.T) {
	log.Println("TestExclusive running.")
	tempFileName, err := createTemp()
	if err != nil {
		t.Fatal(err)
	}

	// Read lock #1
	f, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	lf := New(f)
	if err = lf.RLock(); err != nil {
		t.Fatal(err)
	}

	// Read lock #2
	f2, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	lf2 := New(f2)
	if err = lf2.RLock(); err != nil {
		t.Fatal(err)
	}

	// Exclusive lock
	f3, err := os.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	lfex := New(f3)
	lfex.DisableBlocking()

	// Try exclusive lock (should fail)
	err = lfex.RWLock()
	if err == nil {
		t.Fatal("Expected failure on exclusive lock")
	} else if err != LOCK_CONFLICT {
		t.Fatal(err)
	}

	if err = lf.UnlockAndClose(); err != nil {
		t.Fatal(err)
	}

	// Try exclusive lock (still should fail, another hold read lock)
	err = lfex.RWLock()
	if err == nil {
		t.Fatal("Expected failure on exclusive lock")
	} else if err != LOCK_CONFLICT {
		t.Fatal(err)
	}

	if err = lf2.UnlockAndClose(); err != nil {
		t.Fatal(err)
	}

	// Try exclusive lock (should succeed, all locks released)
	if err = lfex.RWLock(); err != nil {
		t.Fatal(err)
	}
	
	if err = lfex.UnlockAndClose(); err != nil {
		t.Fatal(err)
	}
}