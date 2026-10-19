package app

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/peacock/internal/logs"
	"github.com/dubeyKartikay/peacock/internal/source"
	"github.com/dubeyKartikay/peacock/internal/tui"
)

func readBatch(ctx context.Context, events <-chan source.Event, maxEntries int) tea.Cmd {
	return func() tea.Msg {
		var batch tui.EntriesMsg
		for len(batch.Entries) < maxEntries {
			select {
			case <-ctx.Done():
				return nil
			case event, ok := <-events:
				if !ok {
					batch.Done = true
					return batch
				}
				switch {
				case event.Line != nil:
					batch.Entries = append(batch.Entries, logs.ParseLine(*event.Line))
				case event.Err != nil:
					batch.Err = event.Err
					return batch
				case event.Done:
					batch.Done = true
					return batch
				}
			default:
				return batch
			}
		}
		return batch
	}
}
