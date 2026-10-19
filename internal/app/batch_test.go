package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dubeyKartikay/peacock/internal/source"
	"github.com/dubeyKartikay/peacock/internal/tui"
)

func lineEvent(line string) source.Event {
	return source.Event{Line: &line}
}

func TestReadBatchHonorsConfiguredLimit(t *testing.T) {
	for _, limit := range []int{1, 3, 256} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			events := make(chan source.Event, limit+1)
			for i := range limit + 1 {
				events <- lineEvent(fmt.Sprintf("line-%d", i))
			}
			read := readBatch(context.Background(), events, limit)
			first := read().(tui.EntriesMsg)
			if len(first.Entries) != limit || first.Done {
				t.Fatalf("expected %d entries before EOF, got %+v", limit, first)
			}
			last := read().(tui.EntriesMsg)
			if len(last.Entries) != 1 || last.Entries[0].Raw != fmt.Sprintf("line-%d", limit) {
				t.Fatalf("unexpected remaining batch: %+v", last)
			}
			for i, entry := range first.Entries {
				if want := fmt.Sprintf("line-%d", i); entry.Raw != want {
					t.Fatalf("entry reordered or overwritten: got %q want %q", entry.Raw, want)
				}
			}
		})
	}
}

func TestReadBatchReturnsSparseInputWithoutEOF(t *testing.T) {
	events := make(chan source.Event, 1)
	events <- lineEvent(`{"message":"quiet stream"}`)
	batch := readBatch(context.Background(), events, 256)().(tui.EntriesMsg)
	if len(batch.Entries) != 1 || batch.Entries[0].Message.Text != "quiet stream" || batch.Done {
		t.Fatalf("unexpected sparse batch: %+v", batch)
	}
}

func TestReadBatchPreservesErrorAndDoneOrder(t *testing.T) {
	errSource := errors.New("source failed")
	events := make(chan source.Event, 4)
	events <- lineEvent("before error")
	events <- source.Event{Err: errSource}
	events <- lineEvent("after error")
	events <- source.Event{Done: true}
	close(events)
	read := readBatch(context.Background(), events, 256)
	first := read().(tui.EntriesMsg)
	last := read().(tui.EntriesMsg)
	if len(first.Entries) != 1 || first.Entries[0].Raw != "before error" || first.Err != errSource || first.Done {
		t.Fatalf("unexpected batch at error boundary: %+v", first)
	}
	if len(last.Entries) != 1 || last.Entries[0].Raw != "after error" || !last.Done || last.Err != nil {
		t.Fatalf("unexpected batch at done boundary: %+v", last)
	}
}

func TestReadBatchFlushesOnChannelClose(t *testing.T) {
	events := make(chan source.Event, 1)
	events <- lineEvent("last line")
	close(events)
	batch := readBatch(context.Background(), events, 256)().(tui.EntriesMsg)
	if len(batch.Entries) != 1 || batch.Entries[0].Raw != "last line" || !batch.Done {
		t.Fatalf("expected final batch on close, got %+v", batch)
	}
}

func TestReadBatchReturnsEmptyWhenNoInputIsReady(t *testing.T) {
	batch := readBatch(context.Background(), make(chan source.Event), 256)().(tui.EntriesMsg)
	if len(batch.Entries) != 0 || batch.Err != nil || batch.Done {
		t.Fatalf("expected empty batch for idle source, got %+v", batch)
	}
}

func TestReadBatchStopsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if msg := readBatch(ctx, make(chan source.Event), 256)(); msg != nil {
		t.Fatalf("expected no message on cancellation, got %v", msg)
	}
}
