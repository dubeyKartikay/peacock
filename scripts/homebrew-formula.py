import argparse
import hashlib
import json
from pathlib import Path
import re


def generate(dist):
    version = json.loads((dist / "metadata.json").read_text())["version"]
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?", version):
        raise ValueError(f"invalid release version: {version}")

    checksums = {}
    for line in (dist / f"peacock_{version}_checksums.txt").read_text().splitlines():
        digest, name = line.split()
        checksums[name] = digest

    lines = [
        "class Peacock < Formula",
        '  desc "Beautiful JSON log viewer for your terminal."',
        '  homepage "https://github.com/dubeyKartikay/peacock"',
        f'  version "{version}"',
        '  license "MIT"',
    ]
    for platform, os_name in [("macos", "Darwin"), ("linux", "Linux")]:
        lines.extend(["", f"  on_{platform} do"])
        for cpu, arch in [("arm", "arm64"), ("intel", "x86_64")]:
            archive = f"peacock_{os_name}_{arch}.tar.gz"
            digest = hashlib.sha256((dist / archive).read_bytes()).hexdigest()
            if checksums.get(archive) != digest:
                raise ValueError(f"checksum mismatch: {archive}")
            lines.extend([
                f"    on_{cpu} do",
                f'      url "https://github.com/dubeyKartikay/peacock/releases/download/v{version}/{archive}"',
                f'      sha256 "{digest}"',
                "    end",
            ])
        lines.append("  end")
    lines.extend([
        "",
        "  def install",
        '    bin.install "peacock"',
        "  end",
        "",
        "  test do",
        '    assert_equal version.to_s, shell_output("#{bin}/peacock version").strip',
        "  end",
        "end",
        "",
    ])
    output = dist / "homebrew" / "peacock.rb"
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text("\n".join(lines))
    print(output)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--dist", type=Path, default=Path("dist"))
    generate(parser.parse_args().dist)
