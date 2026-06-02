package main

import (
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/dubeyKartikay/peacock/internal/logs"
)

func TestGenerateLogProducesValidPeacockJSON(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	now := time.Date(2026, 5, 16, 12, 30, 45, 123456789, time.FixedZone("PDT", -7*60*60))

	line := generateLog(rng, now)
	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("marshal generated log: %v", err)
	}

	entry := logs.ParseLine(string(data))
	if !entry.Parsed {
		t.Fatalf("expected generated log to parse as JSON: %s", data)
	}
	if entry.Level.Text == "" {
		t.Fatal("expected level to be populated")
	}
	if entry.Timestamp.Text != "2026-05-16T19:30:45.123456789Z" {
		t.Fatalf("unexpected timestamp %q", entry.Timestamp.Text)
	}
	if entry.Message.Text == "" {
		t.Fatal("expected message to be populated")
	}
	if entry.Caller.Text == "" {
		t.Fatal("expected caller to be populated")
	}

	for _, fragment := range []string{"service=", "env=prod", "host=", "request_id=", "trace_id=", "span_id=", "region="} {
		if !strings.Contains(entry.Context.Text, fragment) {
			t.Fatalf("expected context to contain %q, got %q", fragment, entry.Context.Text)
		}
	}
}
