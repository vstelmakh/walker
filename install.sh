#!/usr/bin/env bash
#
# Install walker: download the binary for this OS and architecture from a GitHub release.
#
# Usage: install.sh [version]
#   version      Release version, for example 1.2.3. Default: latest release.
#
# Environment:
#   INSTALL_DIR       Where to put the binary. Default: /usr/local/bin
#   WALKER_REPO_URL   Repository URL. Default: https://github.com/vstelmakh/walker

# Must stay before any bash-only syntax, so "curl ... | sh" fails with a clear message.
if [ -z "${BASH_VERSION:-}" ]; then
    echo "Error: This script needs bash. Run it with: bash install.sh" >&2
    exit 1
fi

set -euo pipefail

main() {
    local repo_url="${WALKER_REPO_URL:-https://github.com/vstelmakh/walker}"
    local install_dir="${INSTALL_DIR:-/usr/local/bin}"
    local version="${1:-latest}"
    local os arch asset base_url tmp_dir expected target sudo_cmd

    if [[ $# -gt 1 ]]; then
        fail "Too many arguments. Usage: install.sh [version]"
    fi

    if [[ "$version" != "latest" && ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        fail "Invalid version '$version'. Expected format X.Y.Z, for example 1.2.3."
    fi

    case "$(uname -s)" in
        Darwin) os="macos" ;;
        Linux) os="linux" ;;
        *) fail "Unsupported OS '$(uname -s)'. Download the binary manually: $repo_url/releases" ;;
    esac

    case "$(uname -m)" in
        arm64 | aarch64) arch="arm64" ;;
        x86_64 | amd64) arch="amd64" ;;
        *) fail "Unsupported architecture '$(uname -m)'. Download the binary manually: $repo_url/releases" ;;
    esac

    asset="walker-$os-$arch"

    if [[ "$version" == "latest" ]]; then
        base_url="$repo_url/releases/latest/download"
    else
        base_url="$repo_url/releases/download/$version"
    fi

    tmp_dir="$(mktemp -d)"
    # Double quotes on purpose: expand now, because tmp_dir is local and gone when the trap runs.
    trap "rm -rf '$tmp_dir'" EXIT

    echo "Downloading $asset ($version)"
    download "$base_url/$asset" "$tmp_dir/$asset" || fail "Download failed: $base_url/$asset"
    download "$base_url/SHA256SUMS" "$tmp_dir/SHA256SUMS" || fail "Download failed: $base_url/SHA256SUMS"

    expected="$(awk -v asset="$asset" '$2 == asset { print $1 }' "$tmp_dir/SHA256SUMS")"
    [[ -n "$expected" ]] || fail "No checksum for $asset in SHA256SUMS."
    [[ "$(sha256 "$tmp_dir/$asset")" == "$expected" ]] || fail "Checksum mismatch for $asset."

    target="$install_dir/walker"

    sudo_cmd=""
    if [[ ! -w "$install_dir" ]]; then
        sudo_cmd="sudo"
        echo "$install_dir is not writable by you. Using sudo to install there, it may ask for your password."
    fi

    $sudo_cmd mkdir -p "$install_dir"
    $sudo_cmd install -m 755 "$tmp_dir/$asset" "$target"

    echo "Installed: $target"
    "$target" --version || fail "Installed, but \"$target --version\" failed."
}

fail() {
    echo "Error: $*" >&2
    exit 1
}

# Usage: download <url> <output file>
download() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$1" -o "$2"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
    else
        fail "curl or wget is required."
    fi
}

# Usage: sha256 <file>
sha256() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | cut -d ' ' -f 1
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | cut -d ' ' -f 1
    else
        fail "sha256sum or shasum is required to verify the download."
    fi
}

main "$@"
