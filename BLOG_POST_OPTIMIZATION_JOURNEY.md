# Peacock: How I Accidentally Benchmarked a 0×0 Terminal (And 4 Other Bruh Moments)

> A story about building a JSON log viewer, chasing bottlenecks, and the humbling realization that your benchmark script is lying to you.

---

## 1. The Problem: Production Is Slow

I built **Peacock**, a terminal JSON log viewer. It looks pretty. It has filters. It has Vim keybindings.

I piped a real production microservice into it.

It was **horribly slow**. Logs stuttered. The UI lagged. I was watching logs in slow motion while the service itself was handling 10k RPS. Embarrassing.

Time to optimize.

---

## 2. Bruh Moment #1: `viewport.SetContent` Does What Now?

I wrote a stress-test benchmark: a synthetic log generator blasting 50,000 lines/sec into Peacock.

CPU profile says the hot path is `viewport.SetContent`.

I dig into the Bubble Tea viewport code. `SetContent` takes a giant string, **splits it on newlines**, and then calls `SetContentLines`.

So it does an **O(n)** pass to build a slice of lines... which I already have as a slice of lines.

**Bruh moment.**

I call `SetContentLines` directly. Skip the string-splitting. Small win. I feel clever.

*Results: barely moved the needle.*

---

## 3. Bruh Moment #2: I'm Rendering the WHOLE ASS BUFFER

I look at `syncViewport`. On every single new log line, I am:

1. Taking the entire buffer.
2. Rendering **every single entry** through Lipgloss.
3. Word-wrapping the full history.
4. Feeding it to the viewport.

Even though the screen can only show ~70 lines.

**BRUH MOMENT.**

I add an optimization: only render the entries that are actually **visible** in the current viewport window. Plus defer live viewport updates until necessary.

**HUGE IMPROVEMENT.**

Throughput jumps ~4× in TTY mode. This is more than enough for my actual use case.

But I am **high on winning** now. I can't stop.

---

## 4. The Ring Buffer: MASSIVE WIN (Headless Mode)

Next profile: allocation costs. Every new log line triggers `append` on a slice. Once the buffer hits its limit, Go copies the entire backing array to shift elements. `runtime.typedslicecopy` is glowing in the CPU profile.

I replace the slice-trimming backlog with a **fixed-size ring buffer**. No more copies on rollover.

In headless mode (no TTY rendering), throughput goes from **9 lines/sec → 6,095 lines/sec**.

That's a **677× speedup**.

**I AM HIGHER ON WINNING.**

---

## 5. Bruh Moment #3: Word Wrapping Is the Final Boss

Back to TTY mode. The next hotspot: `wordwrap`, `ansi.StringWidth`, `go-runewidth`. Basically, Lipgloss figuring out how wide every grapheme is so it can wrap lines to the terminal width.

I think for a long time.

Can I optimize word wrapping? No, Unicode width calculation is a solved-hard problem.

Can I remove it? No, wrapped logs are actually readable.

Then it hits me: **I am recalculating word wrap on every frame for every visible entry.** The log text hasn't changed. The width hasn't changed. Why am I re-rendering?

I cache the **final rendered string** of each entry. Invalidate only when the terminal width changes.

I run the benchmark.

**HUGE IMPROVEMENT??**

**NO HUGE IMPROVEMENT???**

Wait, what? The numbers are almost identical to the ring-buffer run. Suspicious.

I stare at the benchmark script.

---

## 6. Bruh Moment #4: The Terminal Was 0×0

I trace the `script` invocation in `benchmark-stress.sh`.

The `script` command spawns a pseudo-TTY. By default, if you don't explicitly set the dimensions... **it opens at 0 rows by 0 columns**.

So my "TTY" benchmarks for the cache-rendered-entries optimization were running in a **0×0 terminal**.

With zero width, word wrapping is a no-op. There are no visible lines. My cache was caching nothing because there was nothing to render.

**BRUH. MOMENT.**

I fix the script: explicitly `stty rows 70 cols 240` inside the `script` session.

Re-run.

**HUGE IMPROVEMENT.**

In real TTY mode (70×240), the cache takes throughput from ~212 lines/sec to **~651 lines/sec**.

Headless stays at ~6,095 lines/sec (caching doesn't help when there's no rendering).

---

## 7. Summary of the Ladder

| Optimization | TTY (70×240) | Headless | The Lesson |
|---|---|---|---|
| Base | ~52 l/s | ~9 l/s | It's bad. |
| `SetContentLines` shortcut | ~51 l/s | — | Don't micro-optimize what isn't the bottleneck. |
| Visible-only rendering | ~207 l/s | — | Don't render off-screen pixels. |
| Ring buffer | ~212 l/s | **6,095 l/s** | Eliminate copying. |
| Cache rendered entries | **651 l/s** | 6,095 l/s | Don't recompute what hasn't changed. |

---

## 8. Moral of the Story

1. **Profile first.** My first "optimization" (`SetContentLines`) was irrelevant because rendering was the real bottleneck.
2. **The biggest wins are algorithmic.** Rendering only visible lines beat every micro-optimization.
3. **Your benchmark script is part of the system.** I spent time debugging my own cache code when the benchmark was running in a 0×0 terminal. Trust no one — especially not `script -q`.
4. **Know when to stop.** After visible-only rendering, Peacock was already fast enough for production. I only kept going because it was fun. (And because I wanted to see how fast it *could* go.)

The code is all in the repo. The benchmark script is in `scripts/benchmark-stress.sh`. The synthetic log generator is in `testdata/stress-test/`.

Happy logging.
