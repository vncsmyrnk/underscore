SRCDIR = .
DIST = $(SRCDIR)/dist
GO_FILES = $(shell find $(SRCDIR) -type f -name '*.go' -not -path '$(SRCDIR)/result/*' | sort)
COVERAGE_PROFILE = $(abspath $(DIST)/coverage.out)

PREFIX ?= /usr/local
DESTDIR ?=

INSTALL ?= install
INSTALL_PROGRAM = $(INSTALL)
INSTALL_DATA = $(INSTALL) -m 644
INSTALL_DATA_DIR = cp -r

all: build

fmt:
	@if [ -n "$(GO_FILES)" ]; then \
		gofmt -w $(GO_FILES); \
		goimports -w $(GO_FILES); \
	fi

fmt-check:
	@if [ -n "$(GO_FILES)" ]; then \
		diff="$$(gofmt -d $(GO_FILES))"; \
		if [ -n "$$diff" ]; then \
			printf '%s\n' "$$diff"; \
			exit 1; \
		fi; \
		diff="$$(goimports -d $(GO_FILES))"; \
		if [ -n "$$diff" ]; then \
			printf '%s\n' "$$diff"; \
			exit 1; \
		fi; \
	fi

lint:
	golangci-lint run ./...

test:
	go test -race ./...

coverage:
	mkdir -p $(DIST)
	go test -covermode=count -coverprofile=$(COVERAGE_PROFILE) ./...
	UNDERSCORE_COVERAGE_PROFILE=$(COVERAGE_PROFILE) go test ./internal/testsupport/coverage -run TestEnforceThresholdsFromProfile -count=1

shell-test:
	prove -v

build: $(SRCDIR)/underscore $(shell find $(SRCDIR) -type f -name '*.elv')
	go build ./...
	elvish -compileonly $?

check:
	@$(MAKE) fmt-check
	@$(MAKE) lint
	@$(MAKE) test
	@$(MAKE) coverage
	@$(MAKE) shell-test

install: all
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/share/zsh/site-functions
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/share/underscore
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/bin
	$(INSTALL_PROGRAM) $(SRCDIR)/underscore $(DESTDIR)$(PREFIX)/bin
	$(INSTALL_DATA_DIR) $(SRCDIR)/shell $(DESTDIR)$(PREFIX)/share/underscore
	$(INSTALL_DATA_DIR) $(SRCDIR)/scripts $(DESTDIR)$(PREFIX)/share/underscore
	$(INSTALL_DATA) $(SRCDIR)/completions/zsh $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_underscore

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/underscore
	rm -rf $(DESTDIR)$(PREFIX)/share/underscore
	rm -f $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_underscore

.PHONY: all fmt fmt-check lint test coverage shell-test build check install uninstall
