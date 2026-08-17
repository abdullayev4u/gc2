# gc2 — development tasks
#
# The install directory follows the same rule the go tool uses: GOBIN when it
# is set, otherwise GOPATH/bin.
#
# Note "go install ./cmd" is deliberately not used anywhere here: it names the
# binary after its package directory, so it would install "cmd" rather than
# "gc2". Every target builds with an explicit -o.

BINARY := gc2
PKG    := ./cmd

GOBIN_DIR := $(shell go env GOBIN)
ifeq ($(GOBIN_DIR),)
GOBIN_DIR := $(shell go env GOPATH)/bin
endif

TARGET := $(GOBIN_DIR)/$(BINARY)

.DEFAULT_GOAL := help
.PHONY: help build test vet fmt check update4macos update4lnx uninstall status \
	guard-macos guard-linux

help:
	@echo "gc2 development tasks"
	@echo
	@echo "  make update4macos   check, build and install to $(TARGET)"
	@echo "  make update4lnx     same, on Linux"
	@echo "  make status         show the installed binary and where it came from"
	@echo "  make check          gofmt check, vet and tests"
	@echo "  make build          compile everything without producing a binary"
	@echo "  make test           run the test suite"
	@echo "  make vet            run go vet"
	@echo "  make fmt            format the tree in place"
	@echo "  make uninstall      remove $(TARGET)"

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

# gofmt -l lists files that need formatting; anything listed fails the check
# rather than being silently reformatted.
check: build vet
	@unformatted="$$(gofmt -l . )"; \
	if [ -n "$$unformatted" ]; then \
		echo "these files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@$(MAKE) --no-print-directory test

## update4macos installs gc2 into this machine's Go bin directory.
##
## The checks run first on purpose: this target overwrites the gc2 you use
## every day, and a broken build landing there is a bad afternoon.
##
## The binary is built next to its destination and then renamed over it, so the
## replacement is a same-directory rename. That keeps it atomic and avoids
## "text file busy" if a gc2 process happens to be running.
update4macos: guard-macos check
	@mkdir -p "$(GOBIN_DIR)"
	@tmp="$(TARGET).new.$$$$"; \
	if go build -trimpath -o "$$tmp" $(PKG); then \
		mv -f "$$tmp" "$(TARGET)"; \
	else \
		rm -f "$$tmp"; \
		exit 1; \
	fi
	@echo
	@echo "installed: $(TARGET)"
	@echo "version:   $$($(TARGET) version)"
	@command -v $(BINARY) >/dev/null 2>&1 || \
		echo "note: $(BINARY) is not on your PATH; add $(GOBIN_DIR) to it"

## update4lnx is update4macos for Linux. The recipe is deliberately a copy
## rather than a shared target: update4macos is the one people use every day,
## and keeping it exactly as it was is worth more here than removing the
## duplication. Any change to one should be made to the other.
update4lnx: guard-linux check
	@mkdir -p "$(GOBIN_DIR)"
	@tmp="$(TARGET).new.$$$$"; \
	if go build -trimpath -o "$$tmp" $(PKG); then \
		mv -f "$$tmp" "$(TARGET)"; \
	else \
		rm -f "$$tmp"; \
		exit 1; \
	fi
	@echo
	@echo "installed: $(TARGET)"
	@echo "version:   $$($(TARGET) version)"
	@command -v $(BINARY) >/dev/null 2>&1 || \
		echo "note: $(BINARY) is not on your PATH; add $(GOBIN_DIR) to it"

status:
	@echo "install dir: $(GOBIN_DIR)"
	@if [ -x "$(TARGET)" ]; then \
		echo "binary:      $(TARGET)"; \
		echo "built:       $$(date -r "$(TARGET)" '+%Y-%m-%d %H:%M')"; \
		echo "version:     $$($(TARGET) version)"; \
	else \
		echo "binary:      not installed — run 'make update4macos'"; \
	fi
	@echo "config:      $$(go run $(PKG) config path 2>/dev/null || echo unknown)"

uninstall:
	rm -f "$(TARGET)"
	@echo "removed $(TARGET)"

# The macOS folder-icon support in tools/ostools is cgo against Cocoa, so this
# target is genuinely platform-specific rather than just named that way.
guard-macos:
	@[ "$$(uname -s)" = "Darwin" ] || { \
		echo "update4macos targets macOS; this machine reports $$(uname -s)"; \
		exit 1; \
	}

# On Linux the Cocoa file is excluded by its build tag and SetCustomIcon is a
# no-op, so the guard is about running the target you meant, not about cgo.
guard-linux:
	@[ "$$(uname -s)" = "Linux" ] || { \
		echo "update4lnx targets Linux; this machine reports $$(uname -s)"; \
		exit 1; \
	}
