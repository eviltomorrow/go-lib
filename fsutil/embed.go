package fsutil

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func WriteEmbedFSToDisk(embedFS embed.FS, targetDir string) error {
	if err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		if strings.Contains(name, "fs.go") {
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(targetDir, path), 0o755)
		}

		buf, err := embedFS.ReadFile(path)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(filepath.Join(targetDir, path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := file.Write(buf); err != nil {
			return err
		}

		return file.Sync()
	}); err != nil {
		return err
	}
	return nil
}
