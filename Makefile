
SRCS = $(wildcard src/*.go)
BINARIES = wstail
BLDDIR = build
SRCDIR = src

all: $(BINARIES)

$(BINARIES): %: $(BLDDIR)/%

$(BLDDIR)/wstail:
	mkdir -p $(BLDDIR)
	cd $(SRCDIR) && go build -o $(abspath $@)

clean:
	rm -fr $(BLDDIR)

# Targets
.PHONY: install clean all
# Programs
.PHONY: $(BINARIES)

install:

