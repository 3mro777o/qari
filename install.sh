#!/bin/sh
# Install or update qari on Linux and macOS:
#   curl -fsSL https://3mro777o.github.io/qari/install.sh | sh
set -eu

repo=3mro777o/qari
say() { printf '\033[1m%s\033[0m\n' "$*"; }
die() { printf 'qari install: %s\n' "$*" >&2; exit 1; }
sudo_() { if [ "$(id -u)" = 0 ]; then "$@"; else sudo "$@"; fi; }

case "$(uname -s)" in
Darwin)
	# Homebrew brings mpv and keeps both up to date.
	command -v brew >/dev/null || die "install Homebrew first: https://brew.sh"
	brew tap "$repo" "https://github.com/$repo" >/dev/null 2>&1 || true
	brew trust "$repo" >/dev/null 2>&1 || true # newer Homebrew refuses untrusted taps
	if brew list qari >/dev/null 2>&1; then brew upgrade qari || true; else brew install qari; fi
	say "Done. Type: qari"
	exit 0 ;;
Linux) os=linux ;;
*) die "unsupported system $(uname -s); on Windows use install.ps1" ;;
esac

case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) die "unsupported CPU $(uname -m)" ;;
esac

if ! command -v mpv >/dev/null; then
	say "Installing mpv (needs your password)..."
	{ if command -v pacman >/dev/null; then sudo_ pacman -S --needed --noconfirm mpv
	elif command -v apt-get >/dev/null; then sudo_ apt-get update -qq && sudo_ apt-get install -y mpv
	elif command -v dnf >/dev/null; then sudo_ dnf install -y mpv
	elif command -v zypper >/dev/null; then sudo_ zypper install -y mpv
	else false; fi; } || say "Could not install mpv: install it with your package manager, qari needs it to play."
fi

tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest")
tag=${tag##*/}
ver=${tag#v}
file="qari_${ver}_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

say "Downloading qari $ver..."
curl -fsSL -o "$tmp/$file" "https://github.com/$repo/releases/download/$tag/$file"
curl -fsSL -o "$tmp/checksums.txt" "https://github.com/$repo/releases/download/$tag/checksums.txt"
(cd "$tmp" && grep " $file\$" checksums.txt | sha256sum -c --status) || die "checksum mismatch"
tar -xzf "$tmp/$file" -C "$tmp" qari

# /usr/local/bin is on PATH everywhere, so "qari" works in any new terminal.
q=/usr/local/bin/qari
sudo_ install -m 755 "$tmp/qari" "$q"

# Tab completion for each installed shell; a rerun overwrites the same files.
comp() { # shell, file, [sudo_]
	"$q" completion "$1" >"$tmp/comp" && ${3:-} mkdir -p "${2%/*}" && ${3:-} cp "$tmp/comp" "$2"
}
if "$q" --help 2>&1 | grep -q completion; then # older releases would open the player instead
	if command -v bash >/dev/null; then
		comp bash "${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions/qari"
	fi
	if command -v zsh >/dev/null; then
		d=/usr/local/share/zsh/site-functions
		if zsh -fc 'print -l $fpath' | grep -qx "$d"; then
			comp zsh "$d/_qari" sudo_
		else
			line='command -v qari >/dev/null && (( $+functions[compdef] )) && source <(qari completion zsh) # qari Tab completion'
			grep -qxF "$line" "$HOME/.zshrc" 2>/dev/null || printf '\n%s\n' "$line" >>"$HOME/.zshrc"
		fi
	fi
	if command -v fish >/dev/null; then
		comp fish "${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions/qari.fish"
	fi
fi
say "Done: $($q --version). Type: qari (in a new terminal, Tab completes commands and names)"
