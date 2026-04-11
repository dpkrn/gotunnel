# Release helper: tag, push tag, prime the module proxy.
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
	echo "git push origin $$ver"; \
	git push origin "$$ver"; \
	echo "GOPROXY=https://proxy.golang.org,direct go list -m $(MODULE)@$$ver"; \
	GOPROXY=https://proxy.golang.org,direct go list -m "$(MODULE)@$$ver"

# Ignore extra goals like v1.0.2 so `make pkg v1.0.2` does not fail.
%:
	@:
