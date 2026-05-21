---
name: peacock-benchmark-ladder
description: Use when benchmarking Peacock optimization commits, running the stress-test ladder, measuring throughput, comparing pprof bottlenecks, or analyzing optimization commit performance.
---

# Peacock Benchmark Ladder

Use this skill when the user asks to benchmark Peacock optimization commits, rerun the stress-test ladder, compare throughput across `optimization: ` commits, or explain the pprof bottleneck story.

## Goal

Run a reproducible ladder benchmark across the optimization stack:

1. Start from the current optimized branch.
2. Revert all `optimization: ` commits to establish a base.
3. Run the stress benchmark with a clear `BENCHMARK_NAME`.
4. Reapply each optimization one at a time by reverting the revert commit.
5. Run the same benchmark after each rung.
6. Compare throughput and pprof bottlenecks.

## Important Semantics

- Throughput is the primary score, not total CPU samples.
- Total CPU samples can increase when Peacock gets faster because the pipe is backpressured and a faster Peacock drains more log lines.
- Use `processed_lines / elapsed_seconds` from `metrics.txt` as the throughput metric.
- Use pprof to explain bottleneck movement, not to rank throughput by total samples alone.

## Required Benchmark Script Features

The benchmark script should support these environment variables:

```sh
BENCHMARK_NAME=throughput-00-base
RATE=50000
DURATION=30s
STOP_AFTER=30s
SEED=1
OUT_DIR=/optional/full/path
```

The script must:

- Build `target/peacock`.
- Run the stress generator piped into Peacock.
- Use a pseudo-terminal via `script` because Peacock is a TUI.
- Stop Peacock by sending real Ctrl+C after `STOP_AFTER`.
- Produce `metrics.txt`, `peacock.cpu.prof`, `peacock.pprof.txt`, and `peacock.tty.log`.
- Count processed lines with `tee >(wc -l > "$LINE_COUNT")` between the generator and Peacock.
- Record `processed_lines` and `throughput_lines_per_second` in `metrics.txt`.

Expected pipeline shape:

```sh
go run "$ROOT_DIR/testdata/stress-test/main.go" \
  -rate "$RATE" \
  -duration "$DURATION" \
  -seed "$SEED" \
  | tee >(wc -l > "$LINE_COUNT") \
  | TERM=xterm-256color "$PEACOCK_BIN" --cpuprofile "$CPU_PROFILE"
```

Do not change Peacock behavior just to make the benchmark exit cleanly. The benchmark may exit non-zero because Ctrl+C is injected; the artifacts are still valid if the profile and metrics exist.

## CPU Profiling Compatibility

Some base commits may not support `--cpuprofile`. If the fully reverted base lacks the flag, add a benchmark-only commit that restores only CPU profiling instrumentation, not optimization behavior.

Minimal expected CLI behavior:

- Add `--cpuprofile <path>`.
- Create the profile file.
- Start `runtime/pprof.StartCPUProfile` before `app.Run`.
- Defer `pprof.StopCPUProfile()`.

Keep this as an explicit benchmark instrumentation commit on the ladder branch.

## Branch Workflow

Start by inspecting the worktree:

```sh
git status --short --branch
git log --oneline --decorate -30
```

Create an isolated branch:

```sh
git switch -c benchmark-optimization-ladder
```

Identify contiguous optimization commits:

```sh
git log --oneline --decorate -30
```

Example optimization order from older runs:

```text
3490d3e optimization: use SetContentLines instead of SetContent
3fc9aa1 optimization: set Content Lines Conservatively
6ec135d optimization: apply filter on only entries that are visible
6117562 optimization: defer live viewport work
532ab79 optimization: Ring Buffer
f0e5f04 optimization: cache rendered entries
```

Revert all optimization commits newest-to-oldest, one commit at a time:

```sh
git revert --no-edit <newest-optimization>
git revert --no-edit <next-newest-optimization>
...
git revert --no-edit <oldest-optimization>
```

Use separate revert commits so each optimization can be reapplied independently by reverting its revert.

## Running The Ladder

Use identical settings for every rung unless the user asks otherwise.

Default stress settings:

```sh
RATE=50000 DURATION=30s STOP_AFTER=30s SEED=1
```

Long-run settings:

```sh
RATE=50000 DURATION=5m STOP_AFTER=5m SEED=1
```

Base run:

```sh
BENCHMARK_NAME="throughput-00-base" RATE=50000 DURATION=30s STOP_AFTER=30s SEED=1 scripts/benchmark-stress.sh
```

For each optimization, reapply by reverting the corresponding revert commit, then benchmark:

```sh
git revert --no-edit <revert-of-optimization-1>
BENCHMARK_NAME="throughput-01-set-content-lines" RATE=50000 DURATION=30s STOP_AFTER=30s SEED=1 scripts/benchmark-stress.sh

git revert --no-edit <revert-of-optimization-2>
BENCHMARK_NAME="throughput-02-conservative-content-lines" RATE=50000 DURATION=30s STOP_AFTER=30s SEED=1 scripts/benchmark-stress.sh
```

Recommended benchmark names:

```text
throughput-00-base
throughput-01-set-content-lines
throughput-02-conservative-content-lines
throughput-03-visible-filter-only
throughput-04-defer-live-viewport
throughput-05-ring-buffer
throughput-06-cache-rendered-entries
```

For 5-minute runs, prefix with `throughput5m-`:

```text
throughput5m-00-base
throughput5m-01-set-content-lines
throughput5m-02-conservative-content-lines
throughput5m-03-visible-filter-only
throughput5m-04-defer-live-viewport
throughput5m-05-ring-buffer
throughput5m-06-cache-rendered-entries
```

## Extracting Results

Read each `metrics.txt` and extract:

- `elapsed_ms`
- `processed_lines`
- `throughput_lines_per_second`

Use `grep` for quick extraction:

```sh
grep -R "benchmark_name\|elapsed_ms\|processed_lines\|throughput_lines_per_second" results/stress-benchmark-throughput*/metrics.txt
```

Report a table:

```text
Step | Benchmark | Lines | Throughput
```

Also list output directories so the user can inspect artifacts.

## Known Historical Results

Thirty-second run with `RATE=50000`:

```text
throughput-00-base                         1,308 lines      42 lines/s
throughput-01-set-content-lines            1,087 lines      35 lines/s
throughput-02-conservative-content-lines   1,284 lines      42 lines/s
throughput-03-visible-filter-only         69,967 lines   2,330 lines/s
throughput-04-defer-live-viewport         70,645 lines   2,352 lines/s
throughput-05-ring-buffer                103,651 lines   3,423 lines/s
throughput-06-cache-rendered-entries     102,145 lines   3,401 lines/s
```

Five-minute run with `RATE=50000`:

```text
throughput5m-00-base                         3,024 lines      10 lines/s
throughput5m-01-set-content-lines            1,806 lines       5 lines/s
throughput5m-02-conservative-content-lines   3,230 lines      10 lines/s
throughput5m-03-visible-filter-only        692,581 lines   2,308 lines/s
throughput5m-04-defer-live-viewport        699,238 lines   2,328 lines/s
throughput5m-05-ring-buffer              1,036,083 lines   3,453 lines/s
throughput5m-06-cache-rendered-entries   1,020,874 lines   3,402 lines/s
```

## Pprof Bottleneck Story

Use cumulative pprof to explain the optimization sequence:

```sh
go tool pprof -top -cum target/peacock results/stress-benchmark-<name>/peacock.cpu.prof
```

Expected story:

```text
Base: render/wrap bottleneck
01: render/wrap reduced, SetContentLines insertion/memmove exposed
02: insertion/memmove fixed, render/wrap exposed again
03: render/wrap greatly reduced, append/slice/GC exposed
04: append/slice/GC remains, slight throughput gain
05: append/slice/GC fixed by ring buffer, render/wrap exposed again
06: cache has little effect in append-heavy benchmark; may help rerender-heavy workflows
```

Key historical pprof transitions:

Base bottleneck:

```text
model.Update               29.94s
syncViewport               29.94s
contentLines               23.60s
renderEntry                23.42s
WrapHorizontalOverflow     18.14s
wrapString                 17.13s
wordwrap.Write             14.38s
```

After `01`, new bottleneck:

```text
viewport.SetContentLines   17.77s
slices.Insert              16.83s
runtime.typedslicecopy     16.26s
runtime.memmove            13.02s flat
```

After `02`, insertion fixed:

```text
viewport.SetContentLines   17.77s -> 2.36s
slices.Insert              16.83s -> 1.86s
runtime.typedslicecopy     16.26s -> 1.58s
```

After `03`, render/wrap greatly reduced:

```text
renderEntry                27.52s -> 5.81s
WrapHorizontalOverflow     20.86s -> 4.32s
wrapString                 19.72s -> 4.12s
```

After `03`/`04`, append/storage/GC exposed:

```text
model.appendEntry          ~21.6s
runtime.typedslicecopy     ~14.9s
runtime.gcBgMarkWorker     ~12-13s
runtime.growslice          ~7s
runtime.mallocgcLarge      ~6.7s
```

After `05`, ring buffer fixes append/storage/GC:

```text
model.appendEntry          gone from top path
runtime.typedslicecopy     gone from top path
runtime.gcBgMarkWorker     ~12.4s -> ~1.7s
runtime.memmove            ~8.6s -> ~0.8s flat
runtime.growslice          ~7.0s -> ~0.4s
```

After `05`/`06`, next target remains:

```text
WrapHorizontalOverflow -> wrapString -> wordwrap.String
```

## Analysis Rules

- Do not call a step worse only because total CPU samples increased.
- Prefer throughput for ranking.
- Use pprof for explaining where time moves after each optimization.
- If throughput and pprof disagree, verify line counting and pipeline behavior before drawing conclusions.
- Mention benchmark limitations: this append-heavy stress test may not demonstrate cache benefits that only appear during scrolling, resizing, filtering, or repeated re-rendering.

## Final Response Template

Use a concise summary:

```markdown
Ran the full benchmark ladder with `RATE=... DURATION=... STOP_AFTER=... SEED=...`.

| Step | Benchmark | Lines | Throughput |
|---|---:|---:|---:|
...

Main result: ...

Artifacts:
`results/stress-benchmark-...`

Branch state:
`git status --short --branch` output
```
