PREFIX = /usr/local

GO = go
GO_LDFLAGS = -s -w

all: aubun
install: install-bin

aubun:
	$(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/aubun

install-bin: aubun
	install -Dm755 aubun $(DESTDIR)$(PREFIX)/bin/aubun

tests:
	$(GO) test $(GOFLAGS) ./...

clean:
	rm -f aubun

.PHONY: all aubun install install-bin tests clean
