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

func TestGenerateLogIsDeterministicForSeed(t *testing.T) {
	now := time.Date(2026, 5, 16, 0, 0, 0, 0, time.UTC)
	first := generateLog(rand.New(rand.NewSource(42)), now)
	second := generateLog(rand.New(rand.NewSource(42)), now)

	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first log: %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second log: %v", err)
	}

	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("expected same seed to produce same log\nfirst:  %s\nsecond: %s", firstJSON, secondJSON)
	}
}

func TestStatusAndDurationMatchLevel(t *testing.T) {
	rng := rand.New(rand.NewSource(3))

	for i := 0; i < 100; i++ {
		if got := statusFor("error", rng); got < 500 || got > 504 {
			t.Fatalf("expected error status to be 5xx, got %d", got)
		}
		if got := durationFor("error", rng); got < 300 || got >= 5300 {
			t.Fatalf("expected error duration in [300, 5300), got %d", got)
		}
		if got := durationFor("warn", rng); got < 150 || got >= 2650 {
			t.Fatalf("expected warn duration in [150, 2650), got %d", got)
		}
		if got := durationFor("info", rng); got < 5 || got >= 255 {
			t.Fatalf("expected info duration in [5, 255), got %d", got)
		}
	}
}

func TestErrorCodeOnlyRequiredForErrors(t *testing.T) {
	errorRNG := rand.New(rand.NewSource(4))
	if got := errorCode("error", errorRNG); got == "" {
		t.Fatal("expected error level to always include an error code")
	}

	infoRNG := rand.New(rand.NewSource(5))
	for i := 0; i < 100; i++ {
		got := errorCode("info", infoRNG)
		if got == "" {
			continue
		}
		if !contains([]string{"timeout", "connection_reset", "rate_limited", "deadlock", "invalid_state"}, got) {
			t.Fatalf("unexpected error code %q", got)
		}
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
