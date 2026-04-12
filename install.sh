#!/bin/bash
set -euo pipefail

# Release binaries are built by the repo Makefile into mytunnel/ (e.g. mytunnel/mytunnel-linux).
# GitHub release assets use the same basenames (mytunnel-linux, mytunnel-mac, …).
#
# Optional:
#   MYTUNNEL_VERSION=v1.0.6  — pin curl downloads and go install to that tag
#   MYTUNNEL_USE_GO=1        — install with go install instead of curl

echo "Installing mytunnel..."

REPO_URL="https://github.com/dpkrn/gotunnel"
if [[ -n "${MYTUNNEL_VERSION:-}" ]]; then
	DOWNLOAD_BASE="${REPO_URL}/releases/download/${MYTUNNEL_VERSION}"
else
	DOWNLOAD_BASE="${REPO_URL}/releases/latest/download"
fi

# Optional: build from source (needs a published module without replace in mytunnel/go.mod).
if [[ "${MYTUNNEL_USE_GO:-}" == "1" ]] && command -v go >/dev/null 2>&1; then
	echo "MYTUNNEL_USE_GO=1: installing with go install..."
	INSTALL_REF="${MYTUNNEL_VERSION:-latest}"
	GOTOOLCHAIN=auto go install "github.com/dpkrn/gotunnel/mytunnel@${INSTALL_REF}"
	sudo cp "$(go env GOPATH)/bin/mytunnel" /usr/local/bin/mytunnel
	sudo chmod +x /usr/local/bin/mytunnel
	echo "Installed via go install → /usr/local/bin/mytunnel"
else
	OS=$(uname -s)
	ARCH=$(uname -m)

	# Download to a temp path so this script works when run from a repo clone
	# where ./mytunnel/ is a source directory (curl cannot -o into a directory).
	TMPROOT="${TMPDIR:-/tmp}"

	if [[ "$OS" == "Linux" ]]; then
		ASSET="mytunnel-linux"
		OUT="${TMPROOT}/mytunnel-install-$$"
	elif [[ "$OS" == "Darwin" ]]; then
		if [[ "$ARCH" == "arm64" ]]; then
			ASSET="mytunnel-mac-arm64"
		else
			ASSET="mytunnel-mac"
		fi
		OUT="${TMPROOT}/mytunnel-install-$$"
	elif [[ "$OS" == MINGW* ]] || [[ "$OS" == MSYS* ]] || [[ "$OS" == CYGWIN* ]]; then
		ASSET="mytunnel-windows.exe"
		OUT="${TMPROOT}/mytunnel-install-$$.exe"
	else
		echo "Unsupported OS: $OS $ARCH"
		exit 1
	fi

	URL="${DOWNLOAD_BASE}/${ASSET}"

	curl -fSL --progress-bar "$URL" -o "$OUT" </dev/tty

	if [[ "$ASSET" == *.exe ]]; then
		if file "$OUT" | grep -qv 'text'; then
			DEST="${LOCALAPPDATA:-$HOME}/bin"
			mkdir -p "$DEST"
			mv "$OUT" "$DEST/mytunnel.exe"
			chmod +x "$DEST/mytunnel.exe"
			echo "Installed → $DEST/mytunnel.exe (add that folder to PATH if needed)"
		else
			echo "❌ Download failed — file is not a binary:"
			cat "$OUT"
			rm -f "$OUT"
			exit 1
		fi
	elif file "$OUT" | grep -qv 'text'; then
		chmod +x "$OUT"
		sudo mv "$OUT" /usr/local/bin/mytunnel
	else
		echo "❌ Download failed — file is not a binary:"
		cat "$OUT"
		rm -f "$OUT"
		exit 1
	fi

	echo ""
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	echo "  If mytunnel fails with:"
	echo "    dial tcp [::1]:9000: connect: connection refused"
	echo "  the GitHub release binary is older than your source: it still"
	echo "  dials localhost:9000. Current code uses clickly.cv:9000."
	echo ""
	echo "  Fix: rebuild mytunnel from this repo and upload a new release, or run:"
	echo "    go run ./mytunnel/main.go http <port>     (from a clone)"
	echo "  Or after a fixed release:"
	echo "    MYTUNNEL_USE_GO=1 curl -fsSL .../install.sh | bash"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi

echo ""
echo "✅ Installed successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Quick Start"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  1. Start your local server (e.g. port 3000)"
echo ""
echo "  2. Run the tunnel:"
echo "       mytunnel http 3000"
echo ""
echo "  3. You'll get a public URL like:"
echo "       https://abc123.clickly.cv"
echo "     Share it — all traffic is forwarded to your local server."
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
