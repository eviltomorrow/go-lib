package fsutil

import (
	"os"
	"path/filepath"

	"github.com/eviltomorrow/go-lib/system"
	"golang.org/x/sys/unix"
)

func RewriteStderrToFile() error {
	if err := MkdirAll(system.Directory.LogDir()); err != nil {
		return err
	}

	panicFile, err := os.OpenFile(filepath.Join(system.Directory.LogDir(), "panic.log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer panicFile.Close()

	if err = unix.Dup2(int(panicFile.Fd()), int(os.Stderr.Fd())); err != nil {
		return err
	}

	return nil
}
