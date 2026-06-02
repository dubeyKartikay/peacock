package source

import (
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
)

func TestOpenSelectsFileSourceWithoutFollow(t *testing.T) {
	path := writeLogFile(t, "one\n")
	cfg := appconfig.DefaultConfig()

	src, err := Open(path, nil, &cfg)
	if err != nil {
		t.Fatalf("open source: %v", err)
	}
	defer src.Close()

	if _, ok := src.(*fileSource); !ok {
		t.Fatalf("expected file source, got %T", src)
	}
	if cfg.Source.FileFollow {
		t.Fatal("expected file source selection to preserve non-follow mode")
	}
}

func TestOpenSelectsTailedFileSourceWithFollow(t *testing.T) {
	path := writeLogFile(t, "one\n")
	cfg := appconfig.DefaultConfig()
	cfg.Source.FileFollow = true
	cfg.Source.FilePoll = true

	src, err := Open(path, nil, &cfg)
	if err != nil {
		t.Fatalf("open source: %v", err)
	}
	defer src.Close()

	if _, ok := src.(*tailedFileSource); !ok {
		t.Fatalf("expected tailed file source, got %T", src)
	}
}

func TestOpenSelectsStdinSourceForPipeAndEnablesFollowMode(t *testing.T) {
	stdin := writePipe(t, "one\n")
	cfg := appconfig.DefaultConfig()

	src, err := Open("", stdin, &cfg)
	if err != nil {
		t.Fatalf("open source: %v", err)
	}
	defer src.Close()

	if _, ok := src.(*stdinSource); !ok {
		t.Fatalf("expected stdin source, got %T", src)
	}
	if !cfg.Source.FileFollow {
		t.Fatal("expected stdin source to enable follow mode")
	}
}

func TestOpenReturnsUsageErrorWithoutInput(t *testing.T) {
	cfg := appconfig.DefaultConfig()

	if _, err := Open("", nil, &cfg); err == nil {
		t.Fatal("expected usage error without file path or piped stdin")
	}
}

func writeLogFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "app.log")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write log file: %v", err)
	}
	return path
}
