package logs

import (
	"strings"
	"testing"
)

func TestRenderPlainFormatsStructuredLogs(t *testing.T) {
	entry := ParseLine(`{"level":"error","time":"2026-03-25T12:00:00Z","message":"request failed","caller":"api.go:88","status":503,"url":"https://example.com"}`)

	plain := RenderPlain(entry)
	checks := []string{
		"2026-03-25T12:00:00Z ",
		"[ERROR] ",
		"request failed",
		" caller=api.go:88",
		" status=503",
		" url=https://example.com",
	}
	for _, check := range checks {
		if !strings.Contains(plain, check) {
			t.Fatalf("expected %q in %q", check, plain)
		}
	}
}

func TestTruncatePartsAddsEllipsis(t *testing.T) {
	parts := []Part{{Kind: PartMessage, Text: "abcdefghijklmnopqrstuvwxyz"}}
	truncated := TruncateParts(parts, 10)
	if got, want := truncated[0].Text, "abcdefghi…"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
