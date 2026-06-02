package source

import (
	"os"
	"strings"
	"testing"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
)

func TestStdinSourceEmitsLinesAndDone(t *testing.T) {
	file := writePipe(t, "one\ntwo\n")

	src := NewStdinSource(file, appconfig.DefaultConfig().Input)
	defer src.Close()

	if got := src.Name(); got != "stdin" {
		t.Fatalf("expected stdin source name, got %q", got)
	}
	for _, want := range []string{"one", "two"} {
		if got := nextLine(t, src.Events()); got != want {
			t.Fatalf("expected line %q, got %q", want, got)
		}
	}
	expectDone(t, src.Events())
}

func TestStdinSourceReportsScannerLimitErrors(t *testing.T) {
	file := writePipe(t, "aaaaaa\n")
	cfg := appconfig.DefaultConfig().Input
	cfg.ScannerInitialBufferBytes = 1
	cfg.ScannerMaxBufferBytes = 4

	src := NewStdinSource(file, cfg)
	defer src.Close()

	err := nextError(t, src.Events())
	if !strings.Contains(err.Error(), "token too long") {
		t.Fatalf("expected token too long scanner error, got %v", err)
	}
}

func writePipe(t *testing.T, content string) *os.File {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	if _, err := writer.WriteString(content); err != nil {
		t.Fatalf("write pipe: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	t.Cleanup(func() {
		_ = reader.Close()
	})
	return reader
}
