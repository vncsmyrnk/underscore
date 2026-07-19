SRCDIR = .
DIST = $(SRCDIR)/dist

PREFIX ?= /usr/local
DESTDIR ?=

INSTALL ?= install
INSTALL_PROGRAM = $(INSTALL)
INSTALL_DATA = $(INSTALL) -m 644
INSTALL_DATA_DIR = cp -r

all: $(DIST)

$(DIST): $(SRCDIR)/underscore.elv $(shell find $(SRCDIR)/scripts -type f)
	elvish -compileonly $^
	@mkdir -p $@/bin $@/share/underscore $@/share/zsh/site-functions
	install -m 755 $< $@/bin/underscore
	cp -r $(SRCDIR)/scripts $@/share/underscore
	cp -r $(SRCDIR)/entrypoints $@/share/underscore
	cp $(SRCDIR)/completions/zsh $@/share/zsh/site-functions/_underscore
	@find $@/share/underscore/scripts/ -print0 | xargs -0 chmod 755
	@find $@/share/underscore/entrypoints/ -type d -print0 | xargs -0 chmod 755
	@find $@/share/underscore/entrypoints/ -type f -print0 | xargs -0 chmod 644

clean:
	@rm -rf $(DIST)

install: all
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/share/underscore/scripts
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/share/zsh/site-functions
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/bin
	$(INSTALL_PROGRAM) $(DIST)/bin/underscore $(DESTDIR)$(PREFIX)/bin
	$(INSTALL_DATA_DIR) $(DIST)/share/underscore $(DESTDIR)$(PREFIX)/share
	$(INSTALL_DATA) $(DIST)/share/zsh/site-functions/_underscore $(DESTDIR)$(PREFIX)/share/zsh/site-functions

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/underscore
	rm -rf $(DESTDIR)$(PREFIX)/share/underscore
	rm -f $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_underscore

.PHONY: clean install uninstall
