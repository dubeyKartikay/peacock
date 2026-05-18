package logs

import "testing"

func TestParseLineExtractsCanonicalFields(t *testing.T) {
	entry := ParseLine(`{"level":"warn","time":"2026-03-25T12:00:00Z","message":"disk almost full","caller":"main.go:42","host":"prod-1","retry":3}`)

	if !entry.Parsed {
		t.Fatal("expected JSON log line to parse")
	}
	if entry.Level != "warn" {
		t.Fatalf("expected warn level, got %q", entry.Level)
	}
	if entry.Timestamp != "2026-03-25T12:00:00Z" {
		t.Fatalf("unexpected timestamp %q", entry.Timestamp)
	}
	if entry.Message != "disk almost full" {
		t.Fatalf("unexpected message %q", entry.Message)
	}
	if entry.Caller != "main.go:42" {
		t.Fatalf("unexpected caller %q", entry.Caller)
	}
	if got, want := len(entry.Context), 2; got != want {
		t.Fatalf("expected %d context fields, got %d", want, got)
	}
	if entry.Context[0].Key != "host" || entry.Context[0].Value != "prod-1" {
		t.Fatalf("unexpected first context field %#v", entry.Context[0])
	}
	if entry.Context[1].Key != "retry" || entry.Context[1].Value != "3" {
		t.Fatalf("unexpected second context field %#v", entry.Context[1])
	}
}

func TestParseLineSupportsAliasesAndRawFallback(t *testing.T) {
	aliased := ParseLine(`{"level":"INFO","timestamp":"2026-03-25T12:00:00Z","msg":"hello","file":"app.go:9","request_id":"abc 123"}`)
	if !aliased.Parsed {
		t.Fatal("expected aliased JSON log line to parse")
	}
	if aliased.Level != "info" {
		t.Fatalf("expected normalized info level, got %q", aliased.Level)
	}
	if aliased.Message != "hello" || aliased.Caller != "app.go:9" {
		t.Fatalf("unexpected aliased extraction: %#v", aliased)
	}
	if aliased.Context[0].Value != `"abc 123"` {
		t.Fatalf("expected spaced string to be quoted, got %q", aliased.Context[0].Value)
	}

	raw := ParseLine(`not-json-at-all`)
	if raw.Parsed {
		t.Fatal("expected invalid JSON to remain raw")
	}
	if raw.Raw != "not-json-at-all" {
		t.Fatalf("unexpected raw line %q", raw.Raw)
	}
}
