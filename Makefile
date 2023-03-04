VERSION = `git describe --tags`

PREFIX     = /usr/local
APPPREFIX  = $(PREFIX)/share/applications
ICONPREFIX = $(PREFIX)/share/icons/hicolor

GOFLAGS = -ldflags="-s -w -X main.Version=$(VERSION)" -buildvcs=false
GO = go

all: vinegar

vinegar:
	$(GO) build $(GOFLAGS)

install: vinegar
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar
	install -Dm644 desktop/io.github.vinegarhq.Vinegar.app.desktop $(DESTDIR)$(APPPREFIX)/io.github.vinegarhq.Vinegar.app.desktop
	install -Dm644 desktop/io.github.vinegarhq.Vinegar.player.desktop $(DESTDIR)$(APPPREFIX)/io.github.vinegarhq.Vinegar.player.desktop
	install -Dm644 desktop/io.github.vinegarhq.Vinegar.studio.desktop $(DESTDIR)$(APPPREFIX)/io.github.vinegarhq.Vinegar.studio.desktop
	install -Dm644 icons/16/io.github.vinegarhq.Vinegar.player.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/io.github.vinegarhq.Vinegar.player.png
	install -Dm644 icons/32/io.github.vinegarhq.Vinegar.player.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/io.github.vinegarhq.Vinegar.player.png
	install -Dm644 icons/48/io.github.vinegarhq.Vinegar.player.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/io.github.vinegarhq.Vinegar.player.png
	install -Dm644 icons/64/io.github.vinegarhq.Vinegar.player.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/io.github.vinegarhq.Vinegar.player.png
	install -Dm644 icons/128/io.github.vinegarhq.Vinegar.player.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/io.github.vinegarhq.Vinegar.player.png
	install -Dm644 icons/16/io.github.vinegarhq.Vinegar.studio.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/io.github.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/32/io.github.vinegarhq.Vinegar.studio.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/io.github.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/48/io.github.vinegarhq.Vinegar.studio.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/io.github.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/64/io.github.vinegarhq.Vinegar.studio.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/io.github.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/128/io.github.vinegarhq.Vinegar.studio.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/io.github.vinegarhq.Vinegar.studio.png

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(APPPREFIX)/io.github.vinegarhq.Vinegar.app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/io.github.vinegarhq.Vinegar.player.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/io.github.vinegarhq.Vinegar.studio.desktop
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/io.github.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/io.github.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/io.github.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/io.github.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/io.github.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/io.github.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/io.github.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/io.github.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/io.github.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/io.github.vinegarhq.Vinegar.studio.png

mime:
	xdg-mime default io.github.vinegarhq.Vinegar.player.desktop x-scheme-handler/roblox-player
	xdg-mime default io.github.vinegarhq.Vinegar.studio.desktop x-scheme-handler/roblox-studio

clean:
	rm -f vinegar

.PHONY: all vinegar install uninstall mime clean
