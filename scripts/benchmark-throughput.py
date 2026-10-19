import argparse
import errno
import fcntl
import json
import os
from pathlib import Path
import pty
import re
import select
import signal
import statistics
import struct
import tempfile
import termios
import time


FINAL_MARKER = "PEACOCK_BENCHMARK_FINISHED"
ANSI_CSI = re.compile(rb"\x1b\[[0-?]*[ -/]*[@-~]")


def positive_int(value):
    number = int(value)
    if number <= 0:
        raise argparse.ArgumentTypeError("must be greater than zero")
    return number


def fixture(count):
    lines = []
    for index in range(count):
        entry = {
            "level": "info",
            "time": "2026-10-19T12:00:00Z",
            "message": FINAL_MARKER if index == count - 1 else "request completed",
            "caller": "server/http.go:121",
            "service": "api",
            "env": "prod",
            "host": "prod-api-01",
            "request_id": f"req_{index:016x}",
            "trace_id": f"{index:032x}",
            "span_id": f"{index:016x}",
            "method": "GET",
            "path": "/v1/users",
            "status": 200,
            "duration_ms": 18,
            "user_id": 12345,
            "region": "ap-south-1",
            "extra": {"attempt": 1, "upstream": "postgres", "queue_depth": 100},
        }
        lines.append(json.dumps(entry, separators=(",", ":")).encode() + b"\n")
    return b"".join(lines)


def measure(binary, config, payload, rows, cols, timeout):
    stdin_read_fd, stdin_write_fd = os.pipe()
    started = time.monotonic()
    pid, terminal_fd = pty.fork()
    if pid == 0:
        fcntl.ioctl(1, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))
        os.dup2(stdin_read_fd, 0)
        os.close(stdin_read_fd)
        os.close(stdin_write_fd)
        env = {key: value for key, value in os.environ.items() if not key.startswith("PEACOCK_")}
        env.update(TERM="xterm-256color", TERM_PROGRAM="Apple_Terminal", LINES=str(rows), COLUMNS=str(cols))
        os.execve(binary, [binary, "--config", config], env)

    os.close(stdin_read_fd)
    os.set_blocking(stdin_write_fd, False)
    sent = 0
    output = b""
    elapsed = None
    terminal_open = True
    try:
        while time.monotonic() - started < timeout:
            readers, writers, _ = select.select(
                [terminal_fd] if terminal_open else [],
                [stdin_write_fd] if stdin_write_fd is not None else [],
                [],
                0.01,
            )
            if writers:
                try:
                    sent += os.write(stdin_write_fd, payload[sent : sent + 65536])
                except BrokenPipeError:
                    os.close(stdin_write_fd)
                    stdin_write_fd = None
                if stdin_write_fd is not None and sent == len(payload):
                    os.close(stdin_write_fd)
                    stdin_write_fd = None
            if readers:
                try:
                    data = os.read(terminal_fd, 65536)
                except OSError as error:
                    if error.errno != errno.EIO:
                        raise
                    data = b""
                terminal_open = bool(data)
                output = (output + data)[-262144:]
                if elapsed is None and FINAL_MARKER.encode() in ANSI_CSI.sub(b"", output):
                    elapsed = time.monotonic() - started
                    os.write(terminal_fd, b"\x03")

            finished, status = os.waitpid(pid, os.WNOHANG)
            if finished:
                pid = None
                if os.waitstatus_to_exitcode(status) != 0 or elapsed is None or sent != len(payload):
                    raise RuntimeError(f"viewer failed or final entry was not rendered: {output[-1000:]!r}")
                return elapsed
        raise TimeoutError(f"viewer did not complete within {timeout}s")
    finally:
        if stdin_write_fd is not None:
            os.close(stdin_write_fd)
        os.close(terminal_fd)
        if pid is not None:
            try:
                os.kill(pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            os.waitpid(pid, 0)


def main():
    parser = argparse.ArgumentParser(
        description="Compare stdin throughput with real terminal rendering on macOS or Linux."
    )
    parser.add_argument("binaries", nargs="+", type=lambda value: str(Path(value).resolve()))
    parser.add_argument("--lines", type=positive_int, default=5000)
    parser.add_argument("--runs", type=positive_int, default=3)
    parser.add_argument("--rows", type=positive_int, default=70)
    parser.add_argument("--cols", type=positive_int, default=240)
    parser.add_argument("--timeout", type=positive_int, default=60)
    args = parser.parse_args()
    for binary in args.binaries:
        if not os.access(binary, os.X_OK):
            parser.error(f"binary is not executable: {binary}")

    payload = fixture(args.lines)
    results = {binary: [] for binary in args.binaries}
    with tempfile.TemporaryDirectory(prefix="peacock-throughput-") as temporary:
        config = str(Path(temporary) / "config.yaml")
        Path(config).write_text("buffer:\n  max_entries: 5000\n")
        for trial in range(args.runs):
            binaries = args.binaries if trial % 2 == 0 else list(reversed(args.binaries))
            for binary in binaries:
                elapsed = measure(binary, config, payload, args.rows, args.cols, args.timeout)
                results[binary].append(elapsed)
                print(json.dumps({
                    "binary": binary, "run": trial + 1, "lines": args.lines,
                    "rows": args.rows, "cols": args.cols, "seconds": round(elapsed, 4),
                    "lines_per_second": round(args.lines / elapsed),
                }), flush=True)
        for binary, times in results.items():
            print(json.dumps({
                "binary": binary,
                "median_lines_per_second": round(args.lines / statistics.median(times)),
            }), flush=True)


if __name__ == "__main__":
    main()
