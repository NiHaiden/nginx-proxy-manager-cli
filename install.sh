#!/usr/bin/env bash
set -euo pipefail

APP_NAME="npmctl"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${NPMCTL_BIN_DIR:-$HOME/.local/bin}"
BIN_PATH="$BIN_DIR/$APP_NAME"

GITHUB_REPO="${NPMCTL_GITHUB_REPO:-NiHaiden/nginx-proxy-manager-cli}"
GITHUB_REF="${NPMCTL_GITHUB_REF:-main}"
GO_PACKAGE="github.com/${GITHUB_REPO}/cmd/${APP_NAME}@${GITHUB_REF}"

log() {
  printf "[npmctl-install] %s\n" "$*"
}

fail() {
  printf "[npmctl-install] ERROR: %s\n" "$*" >&2
  exit 1
}

ensure_requirements() {
  command -v go >/dev/null 2>&1 || fail "go is required"
}

install_binary() {
  mkdir -p "$BIN_DIR"
  if [[ -f "$REPO_DIR/go.mod" ]]; then
    log "Building $APP_NAME from local checkout ($REPO_DIR)"
    local version
    local commit
    local build_date
    version="$(tr -d '[:space:]' < "$REPO_DIR/VERSION")"
    commit="$(cd "$REPO_DIR" && git rev-parse --short HEAD 2>/dev/null || echo unknown)"
    build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    (cd "$REPO_DIR" && go build \
      -trimpath \
      -ldflags="-s -w -X github.com/NiHaiden/nginx-proxy-manager-cli/internal/npmctl.Version=$version -X github.com/NiHaiden/nginx-proxy-manager-cli/internal/npmctl.Commit=$commit -X github.com/NiHaiden/nginx-proxy-manager-cli/internal/npmctl.BuildDate=$build_date" \
      -o "$BIN_PATH" \
      ./cmd/npmctl)
  else
    log "Installing $APP_NAME from $GO_PACKAGE"
    GOBIN="$BIN_DIR" go install "$GO_PACKAGE"
  fi
  log "Installed binary: $BIN_PATH"
}

path_contains_bin_dir() {
  [[ ":$PATH:" == *":$BIN_DIR:"* ]]
}

shell_rc_file() {
  case "${SHELL:-}" in
    */zsh) echo "$HOME/.zshrc" ;;
    */bash) echo "$HOME/.bashrc" ;;
    *) echo "$HOME/.profile" ;;
  esac
}

ensure_path_in_shell_config() {
  local rc_file
  rc_file="$(shell_rc_file)"
  local marker="# Added by npmctl installer"
  local line="export PATH=\"$BIN_DIR:\$PATH\""

  if [[ -f "$rc_file" ]] && grep -Fq "$line" "$rc_file"; then
    log "$BIN_DIR already present in $rc_file"
    return
  fi

  {
    echo ""
    echo "$marker"
    echo "$line"
  } >>"$rc_file"

  log "Added $BIN_DIR to PATH in $rc_file"
}

print_finish_message() {
  if path_contains_bin_dir; then
    log "Installation complete. '$APP_NAME' is available now."
  else
    log "Installation complete. Open a new shell, or run:"
    log "  export PATH=\"$BIN_DIR:\$PATH\""
  fi

  log "Try: $APP_NAME --help"
}

main() {
  ensure_requirements
  install_binary
  ensure_path_in_shell_config
  print_finish_message
}

main "$@"
