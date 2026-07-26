SRCDIR = .
DIST = $(SRCDIR)/dist

PREFIX ?= /usr/local
DESTDIR ?=

INSTALL ?= install
INSTALL_PROGRAM = $(INSTALL)
INSTALL_DATA = $(INSTALL) -m 644
INSTALL_DATA_DIR = cp -r

all: $(SRCDIR)/underscore $(shell find $(SRCDIR) -type f -name '*.elv')
	elvish -compileonly $?

check:
	@prove -v

install: all
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/share/zsh/site-functions
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/bin
	$(INSTALL_PROGRAM) $(SRCDIR)/underscore $(DESTDIR)$(PREFIX)/bin
	$(INSTALL_DATA_DIR) $(SRCDIR)/entrypoints $(DESTDIR)$(PREFIX)/share
	$(INSTALL_DATA_DIR) $(SRCDIR)/scripts $(DESTDIR)$(PREFIX)/share
	$(INSTALL_DATA) $(SRCDIR)/entrypoints/zsh $(DESTDIR)$(PREFIX)/share/zsh/site-functions

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/underscore
	rm -rf $(DESTDIR)$(PREFIX)/share/underscore
	rm -f $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_underscore

.PHONY: check install uninstall
