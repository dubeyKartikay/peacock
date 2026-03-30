#!/usr/bin/env bash
set -euo pipefail

REPO="dubeyKartikay/peacock"
BIN_DIR="${HOME}/.local/bin"
BIN_NAME="peacock"

# ── helpers ──────────────────────────────────────────────────────────────────

info()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
ok()    { printf '\033[1;32m  ✓\033[0m %s\n' "$*"; }
die()   { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

# ── platform detection ───────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "Darwin" ;;
    Linux)  echo "Linux"  ;;
    *)      die "Unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)   echo "x86_64" ;;
    aarch64|arm64)  echo "arm64"  ;;
    i386|i686)      echo "i386"   ;;
    *)              die "Unsupported architecture: $(uname -m)" ;;
  esac
}

# ── shell config detection ────────────────────────────────────────────────────

detect_shell_config() {
  local shell_name
  shell_name="$(basename "${SHELL:-}")"

  case "$shell_name" in
    zsh)
      echo "${ZDOTDIR:-$HOME}/.zshrc"
      ;;
    bash)
      # bash prefers .bash_profile on macOS, .bashrc on Linux
      if [[ "$(uname -s)" == "Darwin" ]]; then
        echo "${HOME}/.bash_profile"
      else
        echo "${HOME}/.bashrc"
      fi
      ;;
    fish)
      echo "${HOME}/.config/fish/config.fish"
      ;;
    *)
      echo "${HOME}/.profile"
      ;;
  esac
}

# ── latest release ────────────────────────────────────────────────────────────

fetch_latest_version() {
  local url="https://api.github.com/repos/${REPO}/releases/latest"
  local version

  if command -v curl &>/dev/null; then
    version="$(curl -fsSL "$url" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  elif command -v wget &>/dev/null; then
    version="$(wget -qO- "$url" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  else
    die "curl or wget is required"
  fi

  [[ -n "$version" ]] || die "Could not determine latest version"
  echo "$version"
}

download() {
  local url="$1" dest="$2"
  if command -v curl &>/dev/null; then
    curl -fsSL -o "$dest" "$url"
  else
    wget -qO "$dest" "$url"
  fi
}

# ── idempotency check ─────────────────────────────────────────────────────────

already_installed() {
  local target_version="$1"
  local shell_config="$2"

  # Binary must exist and report the right version
  if [[ ! -x "${BIN_DIR}/${BIN_NAME}" ]]; then
    return 1
  fi

  local installed_version
  installed_version="$("${BIN_DIR}/${BIN_NAME}" version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)"
  if [[ "$installed_version" != "$target_version" ]]; then
    return 1
  fi

  # PATH must already contain BIN_DIR in the running shell
  if [[ ":${PATH}:" != *":${BIN_DIR}:"* ]]; then
    return 1
  fi

  # Shell config must already export BIN_DIR
  if [[ -f "$shell_config" ]] && ! grep -qF "$BIN_DIR" "$shell_config"; then
    return 1
  fi

  return 0
}

# ── PATH setup ────────────────────────────────────────────────────────────────

add_to_path() {
  local shell_config="$1"
  local shell_name
  shell_name="$(basename "${SHELL:-}")"

  if [[ ":${PATH}:" == *":${BIN_DIR}:"* ]] \
    && [[ -f "$shell_config" ]] \
    && grep -qF "$BIN_DIR" "$shell_config"; then
    ok "${BIN_DIR} already in PATH"
    return
  fi

  info "Adding ${BIN_DIR} to PATH in ${shell_config}"

  mkdir -p "$(dirname "$shell_config")"

  if [[ "$shell_name" == "fish" ]]; then
    printf '\n# peacock\nfish_add_path "%s"\n' "$BIN_DIR" >> "$shell_config"
  else
    printf '\n# peacock\nexport PATH="%s:$PATH"\n' "$BIN_DIR" >> "$shell_config"
  fi

  ok "Added to ${shell_config} — restart your shell or run: source ${shell_config}"
}

# ── main ──────────────────────────────────────────────────────────────────────

# Global so the EXIT trap can always reference it
_tmpdir=""
trap '[[ -n "$_tmpdir" ]] && rm -rf "$_tmpdir"' EXIT

main() {
  local os arch version asset_name download_url shell_config

  os="$(detect_os)"
  arch="$(detect_arch)"
  shell_config="$(detect_shell_config)"
  version="$(fetch_latest_version)"

  info "Installing ${BIN_NAME} ${version} (${os}/${arch})"

  if already_installed "$version" "$shell_config"; then
    ok "${BIN_NAME} ${version} is already installed and PATH is configured — nothing to do"
    exit 0
  fi

  asset_name="${BIN_NAME}_${os}_${arch}.tar.gz"
  download_url="https://github.com/${REPO}/releases/download/${version}/${asset_name}"

  _tmpdir="$(mktemp -d)"

  info "Downloading ${asset_name}"
  download "$download_url" "${_tmpdir}/${asset_name}"

  info "Extracting"
  tar -xzf "${_tmpdir}/${asset_name}" -C "$_tmpdir"

  mkdir -p "$BIN_DIR"
  # The binary inside the archive is just named peacock (no version suffix)
  local extracted_bin
  extracted_bin="$(find "$_tmpdir" -maxdepth 2 -type f -name "$BIN_NAME" | head -1)"
  [[ -n "$extracted_bin" ]] || die "Binary not found inside archive"
  install -m 755 "$extracted_bin" "${BIN_DIR}/${BIN_NAME}"

  ok "Installed to ${BIN_DIR}/${BIN_NAME}"

  add_to_path "$shell_config"

  echo ""
  ok "${BIN_NAME} ${version} installed successfully!"
  if [[ ":${PATH}:" != *":${BIN_DIR}:"* ]]; then
    echo "   Run: source ${shell_config}"
    echo "   Or open a new terminal, then try: ${BIN_NAME} --help"
  else
    echo "   Try it: ${BIN_NAME} --help"
  fi
}

main "$@"
