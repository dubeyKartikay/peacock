package logs

import (
	"strings"
	"testing"
)

func TestParseLineExtractsCanonicalFields(t *testing.T) {
	entry := ParseLine(`{"level":"warn","time":"2026-03-25T12:00:00Z","message":"disk almost full","caller":"main.go:42","host":"prod-1","retry":3}`)

	if !entry.Parsed {
		t.Fatal("expected JSON log line to parse")
	}
	if got := entry.Level.Text; got != "warn" {
		t.Fatalf("expected warn level, got %q", got)
	}
	if got := entry.Timestamp.Text; got != "2026-03-25T12:00:00Z" {
		t.Fatalf("unexpected timestamp %q", got)
	}
	if got := entry.Message.Text; got != "disk almost full" {
		t.Fatalf("unexpected message %q", got)
	}
	if got := entry.Caller.Text; got != "main.go:42" {
		t.Fatalf("unexpected caller %q", got)
	}
	for _, fragment := range []string{"host=prod-1", "retry=3"} {
		if !strings.Contains(entry.Context.Text, fragment) {
			t.Fatalf("expected context to contain %q, got %q", fragment, entry.Context.Text)
		}
	}
}

func TestParseLineSupportsAliasesAndRawFallback(t *testing.T) {
	aliased := ParseLine(`{"level":"info","timestamp":"2026-03-25T12:00:00Z","msg":"hello","file":"app.go:9","request_id":"abc 123"}`)
	if !aliased.Parsed {
		t.Fatal("expected aliased JSON log line to parse")
	}
	if got := aliased.Level.Text; got != "info" {
		t.Fatalf("expected info level, got %q", got)
	}
	if got := aliased.Message.Text; got != "hello" {
		t.Fatalf("unexpected aliased message %q", got)
	}
	if got := aliased.Caller.Text; got != "app.go:9" {
		t.Fatalf("unexpected aliased caller %q", got)
	}
	if !strings.Contains(aliased.Context.Text, `request_id="abc 123"`) {
		t.Fatalf("expected quoted request id, got %q", aliased.Context.Text)
	}

	raw := ParseLine(`not-json-at-all`)
	if raw.Parsed {
		t.Fatal("expected invalid JSON to remain raw")
	}
	if raw.Raw != "not-json-at-all" {
		t.Fatalf("unexpected raw line %q", raw.Raw)
	}
}
