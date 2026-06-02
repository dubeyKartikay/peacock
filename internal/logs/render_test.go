package logs

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestWrapHorizontalOverflowKeepsLinesWithinWidth(t *testing.T) {
	maxWidth := 60
	got := WrapHorizontalOverflow("2026-03-25T12:00:00Z [error] ", "request failed caller=api.go:88 status=503", maxWidth)

	for _, fragment := range []string{
		"2026-03-25T12:00:00Z [error]",
		"request failed",
		"caller=api.go:88",
		"status=503",
	} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected %q in %q", fragment, got)
		}
	}

	for _, line := range strings.Split(got, "\n") {
		if width := lipgloss.Width(line); width > maxWidth {
			t.Fatalf("expected line width <= %d, got %d for %q in %q", maxWidth, width, line, got)
		}
	}
}

func TestWrapHorizontalOverflowKeepsSingleLineWhenContentFits(t *testing.T) {
	got := WrapHorizontalOverflow("2026-03-25T12:00:00Z [info] ", "ok", 80)

	if strings.Contains(got, "\n") {
		t.Fatalf("expected single-line output, got %q", got)
	}
	if got != "2026-03-25T12:00:00Z [info] ok" {
		t.Fatalf("unexpected wrapped output %q", got)
	}
}
