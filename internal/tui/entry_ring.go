package tui

import "github.com/dubeyKartikay/peacock/internal/logs"

type entryRing struct {
	entries []logs.Entry
	start   int
	size    int
}

func newEntryRing(capacity int) entryRing {
	return entryRing{entries: make([]logs.Entry, capacity)}
}

func (r entryRing) Len() int {
	return r.size
}

func (r *entryRing) Reset() {
	r.start = 0
	r.size = 0
}

func (r *entryRing) Append(entry logs.Entry) {
	if len(r.entries) == 0 {
		return
	}

	if r.size == len(r.entries) {
		r.entries[r.start] = entry
		r.start = (r.start + 1) % len(r.entries)
		return
	}

	index := (r.start + r.size) % len(r.entries)
	r.entries[index] = entry
	r.size++
}

func (r *entryRing) AppendBatch(entries []logs.Entry) {
	if len(entries) == 0 || len(r.entries) == 0 {
		return
	}

	if len(entries) >= len(r.entries) {
		entries = entries[len(entries)-len(r.entries):]
		copy(r.entries, entries)
		r.start = 0
		r.size = len(entries)
		return
	}

	for _, entry := range entries {
		r.Append(entry)
	}
}

func (r entryRing) At(index int) *logs.Entry {
	if index < 0 || index >= r.size {
		return nil
	}
	physical := (r.start + index) % len(r.entries)
	return &r.entries[physical]
}

func (r entryRing) Newest(n int) []*logs.Entry {
	if r.size == 0 {
		return nil
	}
	if n <= 0 || n > r.size {
		n = r.size
	}

	start := r.size - n
	entries := make([]*logs.Entry, 0, n)
	for index := start; index < r.size; index++ {
		entries = append(entries, r.At(index))
	}
	return entries
}

func (r entryRing) Range(fn func(*logs.Entry) bool) {
	for index := 0; index < r.size; index++ {
		if !fn(r.At(index)) {
			return
		}
	}
}

func (r entryRing) ReverseRange(fn func(*logs.Entry) bool) {
	for index := r.size - 1; index >= 0; index-- {
		if !fn(r.At(index)) {
			return
		}
	}
}
