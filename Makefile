PREFIX=/usr/local
DESTDIR=
BINDIR=${PREFIX}/bin
DATADIR=${PREFIX}/share

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

install: $(BINARIES)
	install -m 755 -d ${DESTDIR}/${BINDIR}
	install -m 755 $(BLDDIR)/wstail ${DESTDIR}/${BINDIR}/wstail
	install -m 755 -d ${DESTDIR}${DATADIR}
	install -d ${DESTDIR}${DATADIR}/wstail
	cp -r wstail/view ${DESTDIR}${DATADIR}/wstail
