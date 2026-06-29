package fsutil

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func CalFileSha256Sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file, nest error: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to copy file to sha256, nest error: %v", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

var reg = regexp.MustCompile(`\s+`)

func GetFileSha256Sum(path string) (string, error) {
	if !strings.HasSuffix(path, ".sha256") {
		return "", fmt.Errorf("invalid path, must be suffix with .sha256")
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file, nest error: %w", err)
	}
	return strings.ToLower(reg.ReplaceAllString(string(buf), "")), nil
}
