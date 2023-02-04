PREFIX    = /usr/local
APPPREFIX = $(PREFIX)/share/applications

DESKTOP   = desktop/app.desktop desktop/player.desktop desktop/studio.desktop

all: vinegar

vinegar:
	go build $(GOFLAGS) ./cmd/vinegar

install: vinegar $(DESKTOP)
	install -Dm755 vinegar -t $(DESTDIR)$(PREFIX)/bin
	install -Dm644 desktop/app.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.app.desktop
	install -Dm644 desktop/player.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.player.desktop
	install -Dm644 desktop/studio.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.studio.desktop
	install -Dm644 desktop/vinegar.svg $(DESTDIR)$(PREFIX)/share/icons/hicolor/scalable/apps/com.github.vinegar.svg

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.player.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.studio.desktop
	rm -f $(DESTDIR)/usr/share/icons/hicolor/scalable/apps/com.github.vinegar.svg

clean:
	rm -f vinegar

.PHONY: all vinegar clean install uninstall
