package logs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedactedFixtureStaysParseable(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "lazyspotify_redacted.log")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lines := 0
	redactions := 0
	for scanner.Scan() {
		lines++
		entry := ParseLine(scanner.Text())
		if !entry.Parsed {
			t.Fatalf("expected fixture line %d to parse as JSON", lines)
		}
		for _, field := range entry.Context {
			if field.Key == "access_token" && field.Value == "[REDACTED]" {
				redactions++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	if lines == 0 {
		t.Fatal("expected fixture to contain log lines")
	}
	if redactions == 0 {
		t.Fatal("expected fixture to contain redacted token fields")
	}
}

func TestFixtureSupportsFilteringTerms(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "lazyspotify_redacted.log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	content := strings.ToLower(string(data))
	for _, term := range []string{"requesting health", "invalid_grant", "daemon error"} {
		if !strings.Contains(content, term) {
			t.Fatalf("expected fixture to contain %q", term)
		}
	}
}
