package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type logLine struct {
	Level      string         `json:"level"`
	Time       string         `json:"time"`
	Message    string         `json:"message"`
	Caller     string         `json:"caller"`
	Service    string         `json:"service"`
	Env        string         `json:"env"`
	Host       string         `json:"host"`
	RequestID  string         `json:"request_id"`
	TraceID    string         `json:"trace_id"`
	SpanID     string         `json:"span_id"`
	Method     string         `json:"method,omitempty"`
	Path       string         `json:"path,omitempty"`
	Status     int            `json:"status,omitempty"`
	DurationMS int            `json:"duration_ms,omitempty"`
	UserID     int            `json:"user_id,omitempty"`
	Region     string         `json:"region"`
	Extra      map[string]any `json:"extra,omitempty"`
}

var (
	levels   = []string{"debug", "info", "info", "info", "warn", "error"}
	services = []string{"api", "billing", "checkout", "notifications", "worker", "auth"}
	hosts    = []string{"prod-api-01", "prod-api-02", "prod-worker-01", "prod-worker-02", "prod-edge-01"}
	regions  = []string{"us-east-1", "us-west-2", "eu-central-1", "ap-south-1"}
	methods  = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	paths    = []string{"/v1/users", "/v1/orders", "/v1/payments", "/v1/search", "/v1/sessions", "/healthz"}
	callers  = []string{"server/http.go:121", "server/middleware.go:88", "worker/jobs.go:44", "db/query.go:203", "auth/session.go:67"}
	messages = map[string][]string{
		"debug": {"cache lookup completed", "feature flag evaluated", "background job heartbeat"},
		"info":  {"request completed", "job processed", "session refreshed", "payment authorized"},
		"warn":  {"slow request detected", "retrying upstream call", "queue depth above threshold"},
		"error": {"upstream request failed", "database query failed", "payment provider rejected request"},
	}
)

func main() {
	rate := flag.Int("rate", 1000, "log lines to write per second")
	duration := flag.Duration("duration", 0, "how long to run; 0 runs until interrupted")
	seed := flag.Int64("seed", time.Now().UnixNano(), "random seed")
	flag.Parse()

	if *rate <= 0 {
		fmt.Fprintln(os.Stderr, "rate must be greater than zero")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *duration)
		defer cancel()
	}

	rng := rand.New(rand.NewSource(*seed))
	encoder := json.NewEncoder(os.Stdout)
	interval := time.Second / time.Duration(*rate)
	if interval <= 0 {
		interval = time.Nanosecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := encoder.Encode(generateLog(rng, now)); err != nil {
				fmt.Fprintf(os.Stderr, "write log: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func generateLog(rng *rand.Rand, now time.Time) logLine {
	level := pick(rng, levels)
	line := logLine{
		Level:     level,
		Time:      now.UTC().Format(time.RFC3339Nano),
		Message:   pick(rng, messages[level]),
		Caller:    pick(rng, callers),
		Service:   pick(rng, services),
		Env:       "prod",
		Host:      pick(rng, hosts),
		RequestID: fmt.Sprintf("req_%016x", rng.Uint64()),
		TraceID:   fmt.Sprintf("%032x", rng.Uint64()),
		SpanID:    fmt.Sprintf("%016x", rng.Uint64()),
		Region:    pick(rng, regions),
	}

	if rng.Intn(100) < 85 {
		line.Method = pick(rng, methods)
		line.Path = pick(rng, paths)
		line.Status = statusFor(level, rng)
		line.DurationMS = durationFor(level, rng)
		line.UserID = 1000 + rng.Intn(900000)
	}

	if level == "warn" || level == "error" || rng.Intn(100) < 15 {
		line.Extra = map[string]any{
			"attempt":       1 + rng.Intn(4),
			"upstream":      pick(rng, []string{"postgres", "redis", "stripe", "sendgrid", "search"}),
			"sampled":       rng.Intn(100) < 20,
			"queue_depth":   rng.Intn(5000),
			"error_code":    errorCode(level, rng),
			"deployment_id": fmt.Sprintf("deploy-%06d", rng.Intn(1000000)),
		}
	}

	return line
}

func pick(rng *rand.Rand, values []string) string {
	return values[rng.Intn(len(values))]
}

func statusFor(level string, rng *rand.Rand) int {
	switch level {
	case "error":
		return pickInt(rng, []int{500, 502, 503, 504})
	case "warn":
		return pickInt(rng, []int{200, 202, 400, 409, 429})
	default:
		return pickInt(rng, []int{200, 200, 200, 201, 204, 304})
	}
}

func durationFor(level string, rng *rand.Rand) int {
	switch level {
	case "error":
		return 300 + rng.Intn(5000)
	case "warn":
		return 150 + rng.Intn(2500)
	default:
		return 5 + rng.Intn(250)
	}
}

func errorCode(level string, rng *rand.Rand) string {
	if level != "error" && rng.Intn(100) >= 25 {
		return ""
	}
	return pick(rng, []string{"timeout", "connection_reset", "rate_limited", "deadlock", "invalid_state"})
}

func pickInt(rng *rand.Rand, values []int) int {
	return values[rng.Intn(len(values))]
}
