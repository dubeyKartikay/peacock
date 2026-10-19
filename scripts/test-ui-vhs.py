import argparse
import itertools
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import signal
import subprocess
import sys
import time


ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "results" / "vhs-ui"
SEPARATOR = "─" * 80 + "\n"
RECORD = re.compile(r"seq_(\d{6}) (keep|drop) (api|worker)")


def generate_logs(count, rate):
    signal.signal(signal.SIGINT, lambda *_: sys.exit(0))
    signal.signal(signal.SIGTERM, lambda *_: sys.exit(0))
    for index in range(count) if count else itertools.count():
        keep = index % 2 == 0
        service = "api" if index % 4 < 2 else "worker"
        tag = " [ok]" if keep and service == "api" else ""
        entry = {
            "level": "error" if keep else "info",
            "time": "12:00:00",
            "message": f"seq_{index + 1:06d} {'keep' if keep else 'drop'} {service}{tag}",
            "caller": "ui.go:42",
        }
        try:
            print(json.dumps(entry), flush=True)
        except BrokenPipeError:
            os._exit(0)
        if rate:
            time.sleep(1 / rate)


class Tape:
    def __init__(self, name):
        self.name = name
        self.lines = [
            f"Output results/vhs-ui/{name}.gif",
            f"Output results/vhs-ui/{name}.ascii",
            'Set Shell "bash"',
            "Set Width 1400",
            "Set Height 720",
            "Set FontSize 16",
            "Set Framerate 8",
            "Set TypingSpeed 0",
            "Set CursorBlink false",
            "Set WaitTimeout 8s",
            'Env PS1 "vhs> "',
            'Env TERM_PROGRAM "Apple_Terminal"',
        ]
        self.checkpoints = {}
        self.command("export PS1='vhs> '")
        self.command(shlex.join(["cd", str(OUTPUT)]))
        self.command("clear")

    def add(self, *commands):
        self.lines.extend(commands)

    def type(self, text):
        self.add("Type " + json.dumps(text))

    def command(self, command):
        self.type(command)
        self.add("Enter", "Sleep 200ms")

    def capture(self, label):
        path = f"results/vhs-ui/{self.name}-{label}.png"
        self.add("Sleep 300ms", f"Screenshot {path}", "Sleep 200ms")
        self.checkpoints[path] = label

    def filter(self, query):
        self.type("/")
        self.add("Sleep 150ms")
        self.type(query)
        self.add("Enter", "Sleep 500ms")

    def quit(self):
        self.add("Ctrl+C", "Wait+Line /vhs>$/")
        self.command("echo STATUS_$?")
        self.add("Wait+Screen /STATUS_0/")
        self.capture("quit")

    def save(self):
        path = OUTPUT / f"{self.name}.tape"
        path.write_text("\n".join(self.lines) + "\n")
        return path

    def read_checkpoints(self):
        frames = (OUTPUT / f"{self.name}.ascii").read_text().split(SEPARATOR)[:-1]
        commands = self.lines[1:]
        commands = [line for line in commands if not line.startswith("Set Shell ")]
        if len(frames) != len(commands):
            raise AssertionError(f"{self.name}: {len(frames)} snapshots for {len(commands)} commands")
        snapshots = {}
        for command, frame in zip(commands, frames):
            if command.startswith("Screenshot "):
                path = command.removeprefix("Screenshot ")
                snapshots[self.checkpoints[path]] = frame
        return snapshots


def generator_command(count=0, rate=8):
    return shlex.join([
        sys.executable, str(Path(__file__).resolve()), "generate",
        "--count", str(count), "--rate", str(rate),
    ])

def peacock_command(*args):
    return shlex.join([str(ROOT / "target" / "peacock"), "--config", "config.yaml", *args])

def static_tape():
    tape = Tape("static")
    for label, args in (("last-three", ("-n", "3")), ("default-ten", ()), ("full-file", ("-n", "200"))):
        tape.command("clear")
        tape.command(peacock_command(*args, "static.log"))
        tape.add("Wait+Line /vhs>$/")
        tape.capture(label)
        tape.command("echo STATUS_$?")
        tape.add("Wait+Screen /STATUS_0/")
    return tape


def live_tape(name, rate):
    tape = Tape(name)
    command = generator_command(rate=rate) + " | " + peacock_command() if name == "pipe" else peacock_command("-f", "-n", "10", "tail.log")
    tape.command(command)
    tape.add("Wait+Screen /seq_[0-9]{6}/", "Wait+Screen /LIVE/")
    tape.capture("live-one")
    tape.add("Sleep 1s")
    tape.capture("live-two")
    tape.add("Space", "Wait+Screen /PAUSED/")
    tape.capture("paused-one")
    tape.add("Sleep 2s")
    tape.capture("paused-two")
    tape.add("Space", "Wait+Screen /LIVE/", "Sleep 500ms")
    tape.capture("resumed")

    tape.type("/")
    tape.add("Sleep 150ms")
    tape.type("keepp")
    tape.add("Backspace", "Enter", "Sleep 500ms")
    tape.capture("filtered-one")
    tape.add("Sleep 1s")
    tape.capture("filtered-two")
    tape.filter("api")
    tape.capture("and-one")
    tape.add("Sleep 1s")
    tape.capture("and-two")
    tape.add("Space", "Wait+Screen /PAUSED/")
    tape.capture("filtered-paused-one")
    tape.add("Sleep 1s")
    tape.capture("filtered-paused-two")
    tape.add("Space", "Wait+Screen /LIVE/", "Sleep 500ms")
    tape.capture("filtered-resumed")

    tape.type("/")
    tape.add("Sleep 150ms")
    tape.type("cancelled_query")
    tape.add("Escape", "Sleep 300ms")
    tape.capture("cancelled-query")
    tape.add("Backspace", "Sleep 300ms")
    tape.capture("one-filter")
    tape.add("Backspace", "Sleep 300ms")
    tape.capture("cleared-filters")
    tape.filter("never_matches_ui")
    tape.capture("no-results-one")
    tape.add("Sleep 1s")
    tape.capture("no-results-two")
    tape.add("Backspace", "Sleep 300ms")
    tape.capture("restored")
    tape.filter("[ok]")
    tape.capture("literal-filter")
    tape.add("Backspace", "Sleep 300ms")

    tape.add("Space", "Wait+Screen /PAUSED/", "Sleep 200ms")
    tape.capture("navigation-bottom")
    tape.filter("keep")
    tape.capture("paused-filter-edit")
    tape.add("Backspace", "Sleep 300ms")
    tape.capture("paused-filter-cleared")
    tape.type("g")
    tape.add("Sleep 200ms")
    tape.capture("navigation-top")
    for command, label in (
        ("PageDown", "page-down"), ("PageUp", "page-up"),
        ("Ctrl+F", "ctrl-f"), ("Ctrl+B", "ctrl-b"),
        ('Type "G"', "jump-bottom"), ("Up", "arrow-up"), ("Down", "arrow-down"),
    ):
        tape.add(command, "Sleep 200ms")
        tape.capture(label)
    if name == "pipe":
        tape.add("Space", "Wait+Screen /LIVE/")
    tape.quit()
    if name == "pipe":
        tape.command("clear")
        tape.command(generator_command(count=12, rate=25) + " | " + peacock_command())
        tape.add("Wait+Screen /DONE/")
        tape.capture("stdin-eof")
        tape.filter("keep")
        tape.capture("eof-filtered")
        tape.type("/")
        tape.add("Sleep 150ms")
        tape.type("unfinished_query")
        tape.add("Ctrl+C", "Wait+Line /vhs>$/")
        tape.capture("quit-filtering")
    return tape


def records(frame):
    return [
        (int(match[1]), match[2], match[3])
        for line in frame.splitlines() if line.lstrip().startswith("│")
        for match in RECORD.finditer(line)
    ]

def panel(frame):
    return "\n".join(line for line in frame.splitlines() if line.lstrip().startswith(("│", "╭", "╰")))

def check(condition, message, checks):
    if not condition:
        raise AssertionError(message)
    checks.append(message)


def verify_live(name, frames):
    checks = []
    rows = {label: records(frame) for label, frame in frames.items()}
    latest = lambda label: max(row[0] for row in rows[label])
    check(latest("live-two") > latest("live-one"), "new messages appear", checks)
    check(panel(frames["paused-one"]) == panel(frames["paused-two"]), "pause freezes displayed logs", checks)
    check(all("PAUSED" in frames[label] for label in ("paused-one", "paused-two", "filtered-paused-one", "filtered-paused-two")), "paused captures show the PAUSED state", checks)
    check(latest("resumed") > latest("paused-two"), "resume displays queued messages", checks)
    for label in ("filtered-one", "filtered-two", "one-filter", "eof-filtered"):
        if label in rows:
            check(bool(rows[label]) and all(row[1] == "keep" for row in rows[label]), f"{label}: only matching logs", checks)
    check(latest("filtered-two") > latest("filtered-one"), "new messages honor the active filter", checks)
    for label in ("and-one", "and-two", "filtered-resumed", "cancelled-query", "literal-filter"):
        check(bool(rows[label]) and all(row[1:] == ("keep", "api") for row in rows[label]), f"{label}: both predicates hold", checks)
    check(latest("and-two") > latest("and-one"), "new messages honor multiple filters", checks)
    check("cancelled_query" not in frames["cancelled-query"], "Esc cancels query editing and closes the input", checks)
    check(panel(frames["filtered-paused-one"]) == panel(frames["filtered-paused-two"]), "filtered pause freezes logs", checks)
    check(latest("filtered-resumed") > latest("filtered-paused-two"), "filtered resume displays matching queued logs", checks)
    check(any(row[2] == "worker" for row in rows["one-filter"]), "Backspace removes only the last filter", checks)
    check(any(row[1] == "drop" for row in rows["cleared-filters"]), "removing all filters restores other logs", checks)
    check(not rows["no-results-one"] and not rows["no-results-two"], "no-match filter stays empty as input arrives", checks)
    check(bool(rows["restored"]), "removing no-match filter restores logs", checks)
    check(bool(rows["paused-filter-edit"]) and all(row[1] == "keep" for row in rows["paused-filter-edit"]), "filters can be changed while paused", checks)
    check(latest("paused-filter-edit") <= latest("navigation-bottom"), "filtering a paused view does not reveal newer input", checks)
    check(any(row[1] == "drop" for row in rows["paused-filter-cleared"]), "filters can be removed while paused", checks)
    first = 1 if name == "pipe" else min(row[0] for row in rows["live-one"])
    oldest = max(first, latest("navigation-bottom") - 1999)
    check(min(row[0] for row in rows["navigation-top"]) == oldest, "g jumps to oldest retained entry", checks)
    check(latest("page-down") > latest("navigation-top"), "PageDown scrolls down", checks)
    check(rows["page-up"] == rows["navigation-top"], "PageUp returns to top", checks)
    check(latest("ctrl-f") > latest("navigation-top"), "Ctrl+F scrolls down", checks)
    check(rows["ctrl-b"] == rows["navigation-top"], "Ctrl+B returns to top", checks)
    check(rows["jump-bottom"] == rows["navigation-bottom"], "G jumps to newest retained entry", checks)
    check(latest("arrow-up") < latest("jump-bottom"), "Up scrolls one line", checks)
    check(rows["arrow-down"] == rows["jump-bottom"], "Down restores bottom position", checks)
    check("STATUS_0" in frames["quit"], "Ctrl+C exits successfully", checks)
    if name == "pipe":
        check(latest("stdin-eof") == 12 and "DONE" in frames["stdin-eof"], "stdin EOF renders final entry and keeps UI available", checks)
        check("vhs>" in frames["quit-filtering"], "Ctrl+C exits while editing a filter", checks)
    return checks


def run(rate):
    OUTPUT.mkdir(parents=True, exist_ok=True)
    (OUTPUT / "config.yaml").write_text("buffer:\n  max_entries: 2000\ninput:\n  batch_max_entries: 256\n  batch_interval: 33ms\n")
    subprocess.run(["go", "build", "-o", "target/peacock", "./cmd/peacock"], cwd=ROOT, check=True)
    with (OUTPUT / "static.log").open("w") as output:
        subprocess.run(shlex.split(generator_command(count=123, rate=0)), stdout=output, check=True)
    results = {}
    env = {key: value for key, value in os.environ.items() if not key.startswith("PEACOCK_")}
    env.pop("NO_COLOR", None)
    env.update(TERM="xterm-256color", COLORTERM="truecolor")
    terminal_bin = OUTPUT / "bin"
    terminal_bin.mkdir(exist_ok=True)
    ttyd = shutil.which("ttyd")
    if ttyd is None:
        raise RuntimeError("ttyd is required")
    wrapper = terminal_bin / "ttyd"
    wrapper.write_text('#!/bin/sh\nexec ' + shlex.quote(ttyd) + ' -t scrollback=0 "$@"\n')
    wrapper.chmod(0o755)
    env["PATH"] = str(terminal_bin) + os.pathsep + env["PATH"]
    for tape in (static_tape(), live_tape("pipe", rate), live_tape("tail", rate)):
        producer = None
        log_file = None
        if tape.name == "tail":
            log_file = (OUTPUT / "tail.log").open("w")
            producer = subprocess.Popen(shlex.split(generator_command(rate=rate)), stdout=log_file)
        try:
            path = tape.save()
            with (OUTPUT / f"{tape.name}.vhs.log").open("w") as log:
                subprocess.run(["vhs", str(path)], cwd=ROOT, env=env, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=180)
            frames = tape.read_checkpoints()
            (OUTPUT / f"{tape.name}.snapshots.json").write_text(json.dumps(frames, indent=2))
            if tape.name == "static":
                checks = []
                rows = {label: records(frame) for label, frame in frames.items()}
                check([row[0] for row in rows["last-three"]] == [121, 122, 123], "-n 3 renders the last three entries", checks)
                check([row[0] for row in rows["default-ten"]] == list(range(114, 124)), "default static tail renders ten entries", checks)
                check(rows["full-file"][-1][0] == 123, "static file renders the final entry and exits", checks)
            else:
                checks = verify_live(tape.name, frames)
            results[tape.name] = {"status": "passed", "generator_rate": rate, "checks": checks}
            print(f"{tape.name}: {len(checks)} checks passed", flush=True)
        except Exception as error:
            results[tape.name] = {"status": "failed", "error": str(error)}
            (OUTPUT / "report.json").write_text(json.dumps(results, indent=2))
            raise
        finally:
            if producer:
                producer.terminate()
                producer.wait(timeout=5)
                log_file.close()
        (OUTPUT / "report.json").write_text(json.dumps(results, indent=2))
    total = sum(len(result["checks"]) for result in results.values())
    report = [
        "# VHS UI checks", "", f"{total} assertions passed.", "",
        f"Live generator setting: {rate:g} lines/second. Sequence numbers verify that logs continue arriving.", "",
        "Static files render and exit at EOF. Interactive controls are exercised on stdin and followed files.", "",
    ]
    for source, result in results.items():
        report.extend([f"## {source}", "", f"[Recording]({source}.gif)", ""])
        report.extend(f"- {check}" for check in result["checks"])
        report.append("")
    (OUTPUT / "report.md").write_text("\n".join(report))


def main():
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command", required=True)
    run_parser = subparsers.add_parser("run")
    run_parser.add_argument("--rate", type=float, default=8)
    generator = subparsers.add_parser("generate")
    generator.add_argument("--count", type=int, default=0)
    generator.add_argument("--rate", type=float, default=8)
    args = parser.parse_args()
    if args.command == "generate":
        generate_logs(args.count, args.rate)
    else:
        run(args.rate)


if __name__ == "__main__":
    main()
