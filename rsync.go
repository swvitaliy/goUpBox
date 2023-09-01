// Tool gokr-rsync is an rsync receiver Go implementation.
package main

import (
	"github.com/gofrs/flock"
	"goupbox/gokr-rsync/receivermaincmd"
	"log"
	"os"
)

func RsyncMain(args []string, stdin *os.File, stdout *os.File, stderr *os.File) (bool, error) {
	fileLock := flock.New("goupbox.lock")
	locked, err := fileLock.TryLock()
	if err != nil {
		log.Fatal("locking failed: " + err.Error())
		return false, err
	}

	defer fileLock.Unlock()

	if !locked {
		log.Fatal("lock already acquired by another process. Skip rsync")
		return false, nil
	}

	if _, err := receivermaincmd.Main(args, stdin, stdout, stderr); err != nil {
		log.Fatal(err)
		return false, err
	}

	return true, nil
}

//func RsyncMainStr(args []string, stdin string, stdout string, stderr string) {
//	stdinReader := strings.NewReader(stdin)
//	stdoutWriter := bytes.NewBufferString(stdout)
//	stderrWriter := bytes.NewBufferString(stdout)
//	if _, err := receivermaincmd.Main(args, stdinReader, stdoutWriter, stderrWriter); err != nil {
//		log.Fatal(err)
//	}
//}
