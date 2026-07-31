#!/usr/bin/env bash
set -euo pipefail

REPO="hxsggsz/kanba"
INSTALL_DIR="${KANBA_INSTALL_DIR:-$HOME/.local/bin}"

detect_platform() {
  local os arch
  os=$(uname -s)
  arch=$(uname -m)

  case "$os" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *)
      echo "Error: unsupported OS '$os'. Supported: Linux, Darwin." >&2
      exit 1
      ;;
  esac

  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)
      echo "Error: unsupported architecture '$arch'. Supported: amd64, arm64." >&2
      exit 1
      ;;
  esac

  echo "${os} ${arch}"
}

resolve_version() {
  if [ -n "${KANBA_VERSION:-}" ]; then
    echo "$KANBA_VERSION"
    return
  fi

  local latest
  latest=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

  if [ -z "$latest" ]; then
    echo "Error: could not resolve the latest release from GitHub API." >&2
    exit 1
  fi

  echo "$latest"
}

main() {
  local platform
  platform=$(detect_platform) || exit 1
  read -r OS ARCH <<< "$platform"
  VERSION=$(resolve_version)

  if [[ "$VERSION" != v* ]]; then
    VERSION="v${VERSION}"
  fi

  local asset="kanba-${VERSION}-${OS}-${ARCH}.tar.gz"
  local base_url="https://github.com/${REPO}/releases/download/${VERSION}"
  tmp_dir=$(mktemp -d)
  staged_bin=""
  trap 'rm -f "$staged_bin"; rm -rf "$tmp_dir"' EXIT

  echo "Installing kanba ${VERSION} (${OS}/${ARCH})..."

  if ! curl -fsSL -o "${tmp_dir}/${asset}" "${base_url}/${asset}"; then
    echo "Error: release asset not found: ${base_url}/${asset}" >&2
    echo "Check that version '${VERSION}' exists: https://github.com/${REPO}/releases" >&2
    exit 1
  fi

  if ! curl -fsSL -o "${tmp_dir}/checksums.txt" "${base_url}/checksums.txt"; then
    echo "Error: checksums file not found: ${base_url}/checksums.txt" >&2
    echo "Check that version '${VERSION}' exists: https://github.com/${REPO}/releases" >&2
    exit 1
  fi

  echo "Verifying checksum..."
  (cd "$tmp_dir" && grep -F -- "$asset" checksums.txt | sha256sum -c -) || {
    echo "Error: checksum verification failed for ${asset}. Aborting install." >&2
    exit 1
  }

  tar -xzf "${tmp_dir}/${asset}" -C "$tmp_dir"

  mkdir -p "$INSTALL_DIR"

  # Stage the new binary inside INSTALL_DIR itself (not the /tmp-based
  # mktemp dir) so the final `mv` below is a same-filesystem rename, which
  # is atomic. A cross-filesystem mv degrades to copy+unlink and can leave
  # a truncated/corrupt binary in place if interrupted mid-copy.
  staged_bin=$(mktemp "${INSTALL_DIR}/.kanba.XXXXXX")
  cp "${tmp_dir}/kanba-${VERSION}-${OS}-${ARCH}/kanba" "$staged_bin"
  chmod +x "$staged_bin"
  mv "$staged_bin" "${INSTALL_DIR}/kanba"

  echo "Installed kanba ${VERSION} to ${INSTALL_DIR}/kanba"

  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      echo "Warning: ${INSTALL_DIR} is not in your PATH. Add it with:"
      echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac
}

main
