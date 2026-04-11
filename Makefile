# Release helper: tag, push tag, prime the module proxy.
#    # git tag v1.0.1          # or v0.x.y, v2.x.y with /v2 in module path, etc.
#    # git push origin v1.0.1
#
#   make pkg VERSION=v1.0.2
#   make pkg v1.0.2          # same (second word becomes the version; "v" optional)
#
# Commit and push your branch to origin before tagging.

MODULE := github.com/dpkrn/gotunnel

.PHONY: pkg
pkg:
	@set -e; \
	ver="$(VERSION)"; \
	if [ -z "$$ver" ]; then ver="$(word 2,$(MAKECMDGOALS))"; fi; \
	if [ -z "$$ver" ]; then \
		echo "Usage: make pkg VERSION=v1.0.2   or   make pkg v1.0.2" >&2; \
		exit 1; \
	fi; \
	case "$$ver" in v*) ;; *) ver="v$$ver" ;; esac; \
	echo "git tag $$ver"; \
	git tag "$$ver"; \
	echo ""; \
	echo "🫡 version $$ver tagged"; \
	echo ""; \
	echo "git push origin $$ver"; \
	git push origin "$$ver"; \
	echo ""; \
	echo "🫡 version $$ver pushed to origin"; \
	echo ""; \
	echo "GOPROXY=https://proxy.golang.org,direct go list -m $(MODULE)@$$ver"; \
	GOPROXY=https://proxy.golang.org,direct go list -m "$(MODULE)@$$ver"; \
	echo ""; \
	echo ""; \
	echo "https://pkg.go.dev/github.com/dpkrn/gotunnel@$$ver"; \
	echo "redirect to this page and request for indexing."; \
	echo ""

## curl -v github.com/dpkrn/gotunnel@v1.0.3

# Cross-build mytunnel and create a GitHub release with the binaries (outputs under mytunnel/).
# Tag should exist on GitHub first (e.g. `make pkg v1.0.6`).
#
#   make bin VERSION=v1.0.6
#   make bin v1.0.6
#
# Replace assets on an existing release:
#   make release-upload v1.0.6

MYTUNNEL_OUT := mytunnel/mytunnel-mac mytunnel/mytunnel-mac-arm64 mytunnel/mytunnel-linux mytunnel/mytunnel-windows.exe

.PHONY: bin release-upload
bin: $(MYTUNNEL_OUT)
	@set -e; \
	ver="$(VERSION)"; \
	if [ -z "$$ver" ]; then ver="$(word 2,$(MAKECMDGOALS))"; fi; \
	if [ -z "$$ver" ]; then \
		echo "Usage: make bin VERSION=v1.0.6   or   make bin v1.0.6" >&2; \
		exit 1; \
	fi; \
	case "$$ver" in v*) ;; *) ver="v$$ver" ;; esac; \
	notes="Release from commit $$(git rev-parse --short HEAD)."; \
	echo "gh release create $$ver ($(MYTUNNEL_OUT))"; \
	gh release create "$$ver" $(MYTUNNEL_OUT) \
		--title "$$ver" \
		--notes "$$notes"

# With -C mytunnel, -o is relative to mytunnel/ — plain names stay inside mytunnel/.
mytunnel/mytunnel-mac-arm64:
	GOOS=darwin GOARCH=arm64 go build -C mytunnel -a -o mytunnel-mac-arm64 .

mytunnel/mytunnel-mac:
	GOOS=darwin GOARCH=amd64 go build -C mytunnel -a -o mytunnel-mac .

mytunnel/mytunnel-linux:
	GOOS=linux GOARCH=amd64 go build -C mytunnel -a -o mytunnel-linux .

mytunnel/mytunnel-windows.exe:
	GOOS=windows GOARCH=amd64 go build -C mytunnel -a -o mytunnel-windows.exe .

release-upload: $(MYTUNNEL_OUT)
	@set -e; \
	ver="$(VERSION)"; \
	if [ -z "$$ver" ]; then ver="$(word 2,$(MAKECMDGOALS))"; fi; \
	if [ -z "$$ver" ]; then \
		echo "Usage: make release-upload VERSION=v1.0.6   or   make release-upload v1.0.6" >&2; \
		exit 1; \
	fi; \
	case "$$ver" in v*) ;; *) ver="v$$ver" ;; esac; \
	echo "gh release upload $$ver ($(MYTUNNEL_OUT)) --clobber"; \
	gh release upload "$$ver" $(MYTUNNEL_OUT) --clobber

# Ignore extra goals like v1.0.2 so `make pkg v1.0.2` does not fail.
%:
	@:
