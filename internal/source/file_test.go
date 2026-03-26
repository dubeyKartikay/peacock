package source

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	appconfig "peacock/internal/config"
)

func TestTailedFileSourceReceivesAppendedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("existing\n"), 0o600); err != nil {
		t.Fatalf("write seed log: %v", err)
	}

	cfg := appconfig.DefaultConfig()
	cfg.Source.FilePoll = true
	cfg.Source.FileTailLines = 0
	src, err := NewTailedFileSource(path, cfg.Source)
	if err != nil {
		t.Fatalf("new tailed file source: %v", err)
	}
	defer src.Close()

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open log for append: %v", err)
	}
	if _, err := file.WriteString("appended\n"); err != nil {
		t.Fatalf("append log line: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close appended log: %v", err)
	}

	if got := nextLine(t, src.Events()); got != "appended" {
		t.Fatalf("expected appended line, got %q", got)
	}
}

func TestFileSourceReadsOnlyLastConfiguredLinesWithoutFollow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("one\ntwo\nthree\nfour\n"), 0o600); err != nil {
		t.Fatalf("write seed log: %v", err)
	}

	cfg := appconfig.DefaultConfig()
	cfg.Source.FileTailLines = 2

	src, err := NewFileSource(path, cfg.Source)
	if err != nil {
		t.Fatalf("new file source: %v", err)
	}
	defer src.Close()

	if got := nextLine(t, src.Events()); got != "three" {
		t.Fatalf("expected first tailed line, got %q", got)
	}
	if got := nextLine(t, src.Events()); got != "four" {
		t.Fatalf("expected second tailed line, got %q", got)
	}
	expectDone(t, src.Events())
}

func nextLine(t *testing.T, events <-chan Event) string {
	t.Helper()

	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			t.Fatal("timed out waiting for line event")
		case event, ok := <-events:
			if !ok {
				t.Fatal("event stream closed before delivering line")
			}
			switch {
			case event.Line != nil:
				return *event.Line
			case event.Err != nil:
				t.Fatalf("unexpected source error: %v", event.Err)
			}
		}
	}
}

func expectDone(t *testing.T, events <-chan Event) {
	t.Helper()

	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			t.Fatal("timed out waiting for done event")
		case event, ok := <-events:
			if !ok {
				return
			}
			switch {
			case event.Done:
				return
			case event.Err != nil:
				t.Fatalf("unexpected source error: %v", event.Err)
			case event.Line != nil:
				t.Fatalf("unexpected extra line %q", *event.Line)
			}
		}
	}
}
