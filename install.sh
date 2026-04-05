#!/bin/bash

echo "Installing mytunnel..."

OS=$(uname)
ARCH=$(uname -m)

if [ "$OS" = "Linux" ]; then
    URL="https://github.com/DpkRn/gotunnel/releases/latest/download/mytunnel-linux"
elif [ "$OS" = "Darwin" ]; then
    if [ "$ARCH" = "arm64" ]; then
        URL="https://github.com/DpkRn/gotunnel/releases/latest/download/mytunnel-mac-arm64"
    else
        URL="https://github.com/DpkRn/gotunnel/releases/latest/download/mytunnel-mac"
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