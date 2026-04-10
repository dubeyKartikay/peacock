package logs

import (
	"strings"
	"testing"
)

func TestRenderContextJoinsFieldsWithSpaces(t *testing.T) {
	got := renderContext([]Field{
		{Key: "status", Value: "503"},
		{Key: "url", Value: "https://example.com"},
	})

	if got != "status=503 url=https://example.com" {
		t.Fatalf("unexpected rendered context %q", got)
	}
}

func TestWrapHorizontalOverflowKeepsMetadataAndWrapsContent(t *testing.T) {
	metadata := "2026-03-25T12:00:00Z [error] "
	content := "request failed caller=api.go:88 status=503 url=https://example.com"

	got := WrapHorizontalOverflow(metadata, content, 40)
	if !strings.HasPrefix(got, metadata) {
		t.Fatalf("expected wrapped output to preserve metadata prefix, got %q", got)
	}
	if !strings.Contains(got, "\n") {
		t.Fatalf("expected wrapped output to span multiple lines, got %q", got)
	}
	if !strings.Contains(got, "status=503") {
		t.Fatalf("expected wrapped output to preserve content, got %q", got)
	}
}
