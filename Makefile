PREFIX = /usr/local

GO = go
GO_LDFLAGS = -s -w
WCC = x86_64-w64-mingw32-gcc

all: aubun
install: install-aubun

aubun:
	$(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/aubun

robloxmutexer: robloxmutexer.c
	$(WCC) $< -o $@

install-aubun: aubun
	install -Dm755 aubun $(DESTDIR)$(PREFIX)/bin/aubun

install-robloxmutexer: robloxmutexer
	install -Dm755 robloxmutexer.exe $(DESTDIR)$(PREFIX)/bin/robloxmutexer

tests:
	$(GO) test $(GOFLAGS) ./...

clean:
	rm -f aubun robloxmutexer.exe

.PHONY: all install install-aubun install-robloxmutexer tests clean
