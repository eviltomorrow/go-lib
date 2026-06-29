package buildinfo

import (
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestVersion(t *testing.T) {
	color.NoColor = true

	AppName = "personal-api"
	MainVersion = "7.0.5"
	GitSha = "abc123def456"
	BuildTime = "2026-05-25T12:00:00+0800"

	v := Version()
	t.Logf("Version output:\n%s", v)

	if !strings.Contains(v, "Version: ") {
		t.Fatal("expected 'Version: ' in output")
	}
	if !strings.Contains(v, "Runtime: ") {
		t.Fatal("expected 'Runtime: ' in output")
	}
	if !strings.Contains(v, "personal-api") {
		t.Fatal("expected app name in output")
	}
	if !strings.Contains(v, "7.0.5") {
		t.Fatal("expected version in output")
	}
	if !strings.Contains(v, "abc123def456") {
		t.Fatal("expected commit sha in output")
	}
	if !strings.Contains(v, "2026-05-25T12:00:00+0800") {
		t.Fatal("expected build time in output")
	}
}

func TestVersion_EmptyValues(t *testing.T) {
	color.NoColor = true

	AppName = ""
	MainVersion = ""
	GitSha = ""
	BuildTime = ""

	v := Version()
	if !strings.Contains(v, "Version: ") {
		t.Fatal("expected 'Version: ' in output")
	}
	if !strings.Contains(v, "Runtime: ") {
		t.Fatal("expected 'Runtime: ' in output")
	}
}

func TestVersion_ContainsGoInfo(t *testing.T) {
	color.NoColor = true

	v := Version()
	if !strings.Contains(v, GoVersion) {
		t.Fatalf("expected GoVersion %s in output", GoVersion)
	}
	if !strings.Contains(v, GoOSArch) {
		t.Fatalf("expected GoOSArch %s in output", GoOSArch)
	}
}

func TestVersion_Format(t *testing.T) {
	color.NoColor = true

	AppName = "myapp"
	MainVersion = "1.0.0"
	GitSha = "deadbeef"
	BuildTime = "now"

	v := Version()
	lines := strings.Split(strings.TrimSuffix(v, "\r\n"), "\r\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "Version: ") {
		t.Fatalf("first line should start with 'Version: ', got: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "Runtime: ") {
		t.Fatalf("second line should start with 'Runtime: ', got: %s", lines[1])
	}
}

func TestVersion_AppName(t *testing.T) {
	color.NoColor = true

	original := AppName
	defer func() { AppName = original }()

	AppName = "personal-auth"
	v := Version()
	if !strings.Contains(v, "personal-auth") {
		t.Fatal("expected 'personal-auth' in output")
	}
}
