#!/bin/bash
set -euo pipefail

echo "Installing mytunnel..."

# Optional: build from source (needs a published module without replace in mytunnel/go.mod).
if [[ "${MYTUNNEL_USE_GO:-}" == "1" ]] && command -v go >/dev/null 2>&1; then
	echo "MYTUNNEL_USE_GO=1: installing with go install..."
	GOTOOLCHAIN=auto go install github.com/dpkrn/gotunnel/mytunnel@latest
	sudo cp "$(go env GOPATH)/bin/mytunnel" /usr/local/bin/mytunnel
	sudo chmod +x /usr/local/bin/mytunnel
	echo "Installed via go install → /usr/local/bin/mytunnel"
else
	OS=$(uname)
	ARCH=$(uname -m)

	if [ "$OS" = "Linux" ]; then
		URL="https://github.com/dpkrn/gotunnel/releases/latest/download/mytunnel-linux"
	elif [ "$OS" = "Darwin" ]; then
		if [ "$ARCH" = "arm64" ]; then
			URL="https://github.com/dpkrn/gotunnel/releases/latest/download/mytunnel-mac-arm64"
		else
			URL="https://github.com/dpkrn/gotunnel/releases/latest/download/mytunnel-mac"
		fi
	else
		echo "Unsupported OS: $OS $ARCH"
		exit 1
	fi

	curl -fSL --progress-bar "$URL" -o mytunnel </dev/tty

	if file mytunnel | grep -qv 'text'; then
		chmod +x mytunnel
		sudo mv mytunnel /usr/local/bin/
	else
		echo "❌ Download failed — file is not a binary:"
		cat mytunnel
		rm -f mytunnel
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
