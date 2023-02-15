// Tool gokr-rsync is an rsync receiver Go implementation.
package rsync

import (
	"bytes"
	"goupbox/gokr-rsync/receivermaincmd"
	"log"
	"os"
	"strings"
)

func Main(args []string, stdin *os.File, stdout *os.File, stderr *os.File) {
	if _, err := receivermaincmd.Main(args, stdin, stdout, stderr); err != nil {
		log.Fatal(err)
	}
}
func MainStr(args []string, stdin string, stdout string, stderr string) {
	stdinReader := strings.NewReader(stdin)
	stdoutWriter := bytes.NewBufferString(stdout)
	stderrWriter := bytes.NewBufferString(stdout)
	if _, err := receivermaincmd.Main(args, stdinReader, stdoutWriter, stderrWriter); err != nil {
		log.Fatal(err)
	}
}
