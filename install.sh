#!/usr/bin/env bash
set -euo pipefail

APP_NAME="npmctl"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_ROOT="${NPMCTL_HOME:-$HOME/.npmctl}"
VENV_DIR="$INSTALL_ROOT/venv"
BIN_DIR="${NPMCTL_BIN_DIR:-$HOME/.local/bin}"
SHIM_PATH="$BIN_DIR/$APP_NAME"

log() {
  printf "[npmctl-install] %s\n" "$*"
}

fail() {
  printf "[npmctl-install] ERROR: %s\n" "$*" >&2
  exit 1
}

ensure_requirements() {
  command -v python3 >/dev/null 2>&1 || fail "python3 is required"
  [[ -f "$REPO_DIR/pyproject.toml" ]] || fail "pyproject.toml not found in $REPO_DIR"
}

create_venv() {
  log "Creating virtual environment in $VENV_DIR"
  python3 -m venv "$VENV_DIR"
}

install_package() {
  log "Installing $APP_NAME from $REPO_DIR"
  "$VENV_DIR/bin/pip" install --upgrade pip >/dev/null
  "$VENV_DIR/bin/pip" install --upgrade "$REPO_DIR"
}

create_shim() {
  mkdir -p "$BIN_DIR"
  cat >"$SHIM_PATH" <<EOF
#!/usr/bin/env bash
exec "$VENV_DIR/bin/$APP_NAME" "\$@"
EOF
  chmod +x "$SHIM_PATH"
  log "Installed launcher: $SHIM_PATH"
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
  create_venv
  install_package
  create_shim
  ensure_path_in_shell_config
  print_finish_message
}

main "$@"
