PREFIX    = /usr/local
APPPREFIX = $(PREFIX)/share/applications

DESKTOP   = desktop/app.desktop desktop/player.desktop desktop/studio.desktop

all: vinegar

vinegar: main.go util/util.go
	go build -ldflags="-s -w" -o vinegar main.go

install: vinegar $(DESKTOP)
	install -Dm755 vinegar -t $(DESTDIR)$(PREFIX)/bin
	install -Dm644 desktop/app.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.app.desktop
	install -Dm644 desktop/player.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.player.desktop
	install -Dm644 desktop/studio.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.studio.desktop

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.player.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.studio.desktop

clean:
	rm -f vinegar

.PHONY: all vinegar clean install uninstall
