package logs

import (
	"strings"
	"testing"
)

func TestWrapHorizontalOverflowPreservesMetadataAndContent(t *testing.T) {
	got := WrapHorizontalOverflow("2026-03-25T12:00:00Z [error] ", "request failed caller=api.go:88 status=503", 80)

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
}
