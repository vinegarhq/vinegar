PREFIX = /usr/local

GO = go
GO_LDFLAGS = -s -w
WCC = x86_64-w64-mingw32-gcc

all: vinegar
install: install-vinegar

vinegar:
	$(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/vinegar

robloxmutexer: robloxmutexer.c
	$(WCC) $< -o $@

install-vinegar: vinegar
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar

install-robloxmutexer: robloxmutexer
	install -Dm755 robloxmutexer.exe $(DESTDIR)$(PREFIX)/bin/robloxmutexer

tests:
	$(GO) test $(GOFLAGS) ./...

clean:
	rm -f vinegar robloxmutexer.exe

.PHONY: all install install-vinegar install-robloxmutexer tests clean
