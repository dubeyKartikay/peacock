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
	if entry.Level.Kind != PartLevel || entry.Level.Text != "warn" {
		t.Fatalf("unexpected level part %#v", entry.Level)
	}
	if entry.Timestamp.Kind != PartTimestamp || entry.Timestamp.Text != "2026-03-25T12:00:00Z" {
		t.Fatalf("unexpected timestamp part %#v", entry.Timestamp)
	}
	if entry.Message.Kind != PartMessage || entry.Message.Text != "disk almost full" {
		t.Fatalf("unexpected message part %#v", entry.Message)
	}
	if entry.Caller.Kind != PartCaller || entry.Caller.Text != "main.go:42" {
		t.Fatalf("unexpected caller part %#v", entry.Caller)
	}
	if entry.Context.Kind != PartContext || entry.Context.Text != "host=prod-1 retry=3" {
		t.Fatalf("unexpected context part %#v", entry.Context)
	}

	wantSearch := "2026-03-25T12:00:00Z warn disk almost full main.go:42 host=prod-1 retry=3"
	if entry.Search != wantSearch {
		t.Fatalf("unexpected search text %q", entry.Search)
	}
}

func TestParseLineSupportsAliasesAndRawFallback(t *testing.T) {
	aliased := ParseLine(`{"level":"INFO","timestamp":"2026-03-25T12:00:00Z","msg":"hello","file":"app.go:9","request_id":"abc 123"}`)
	if !aliased.Parsed {
		t.Fatal("expected aliased JSON log line to parse")
	}
	if aliased.Level.Text != "INFO" {
		t.Fatalf("expected unmodified INFO level, got %q", aliased.Level.Text)
	}
	if aliased.Message.Text != "hello" || aliased.Caller.Text != "app.go:9" {
		t.Fatalf("unexpected aliased extraction: %#v", aliased)
	}
	if aliased.Context.Text != `request_id="abc 123"` {
		t.Fatalf("expected spaced string to be quoted, got %q", aliased.Context.Text)
	}
	if !strings.Contains(aliased.Search, `request_id="abc 123"`) {
		t.Fatalf("expected search text to include quoted context, got %q", aliased.Search)
	}

	raw := ParseLine(`not-json-at-all`)
	if raw.Parsed {
		t.Fatal("expected invalid JSON to remain raw")
	}
	if raw.Raw != "not-json-at-all" {
		t.Fatalf("unexpected raw line %q", raw.Raw)
	}
	if raw.Search != raw.Raw {
		t.Fatalf("expected raw search text to match raw line, got %q", raw.Search)
	}
}
