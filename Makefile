PREFIX     = /usr/local
APPPREFIX  = $(PREFIX)/share/applications
ICONPREFIX = $(PREFIX)/share/icons/hicolor

FLATPAK = io.github.vinegarhq.Vinegar
VERSION = `git describe --tags --dirty`

GO = go
GO_LDFLAGS = -s -w -X main.Version=$(VERSION)
WCC = x86_64-w64-mingw32-gcc

all: vinegar
install: install-vinegar install-desktop install-icons

vinegar:
	$(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/vinegar

robloxmutexer: robloxmutexer.c
	$(WCC) $< -o $@

install-vinegar: vinegar
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar

install-robloxmutexer: robloxmutexer
	install -Dm755 robloxmutexer.exe $(DESTDIR)$(PREFIX)/bin/robloxmutexer

install-desktop:
	mkdir -p $(DESTDIR)$(APPPREFIX)
	sed "s|\$$FLATPAK|$(FLATPAK)|g" desktop/roblox-app.desktop.in > $(DESTDIR)$(APPPREFIX)/$(FLATPAK).app.desktop
	sed "s|\$$FLATPAK|$(FLATPAK)|g" desktop/roblox-player.desktop.in > $(DESTDIR)$(APPPREFIX)/$(FLATPAK).player.desktop
	sed "s|\$$FLATPAK|$(FLATPAK)|g" desktop/roblox-studio.desktop.in > $(DESTDIR)$(APPPREFIX)/$(FLATPAK).studio.desktop

install-icons:
	install -Dm644 icons/16/roblox-player.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/$(FLATPAK).player.png
	install -Dm644 icons/16/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/$(FLATPAK).studio.png
	install -Dm644 icons/32/roblox-player.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/$(FLATPAK).player.png
	install -Dm644 icons/32/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/$(FLATPAK).studio.png
	install -Dm644 icons/48/roblox-player.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/$(FLATPAK).player.png
	install -Dm644 icons/48/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/$(FLATPAK).studio.png
	install -Dm644 icons/64/roblox-player.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/$(FLATPAK).player.png
	install -Dm644 icons/64/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/$(FLATPAK).studio.png
	install -Dm644 icons/128/roblox-player.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/$(FLATPAK).player.png
	install -Dm644 icons/128/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/$(FLATPAK).studio.png

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(APPPREFIX)/$(FLATPAK).app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/$(FLATPAK).player.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/$(FLATPAK).studio.desktop
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/$(FLATPAK).player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/$(FLATPAK).studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/$(FLATPAK).player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/$(FLATPAK).studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/$(FLATPAK).player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/$(FLATPAK).studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/$(FLATPAK).player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/$(FLATPAK).studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/$(FLATPAK).player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/$(FLATPAK).studio.png

icons: icons/roblox-player.png icons/roblox-studio.png icons/vinegar.svg
	rm -rf icons/16 icons/32 icons/48 icons/64 icons/128
	mkdir  icons/16 icons/32 icons/48 icons/64 icons/128
	convert -background none $^ -resize 16x16   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -background none $^ -resize 32x32   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -background none $^ -resize 48x48   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -background none $^ -resize 64x64   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -background none $^ -resize 128x128 -set filename:f '%w/%t' 'icons/%[filename:f].png'

mime:
	xdg-mime default $(FLATPAK).player.desktop x-scheme-handler/roblox-player
	xdg-mime default $(FLATPAK).player.desktop x-scheme-handler/roblox
	xdg-mime default $(FLATPAK).studio.desktop x-scheme-handler/roblox-studio
	xdg-mime default $(FLATPAK).studio.desktop x-scheme-handler/roblox-studio-auth
	xdg-mime default $(FLATPAK).studio.desktop application/x-roblox-rbxl
	xdg-mime default $(FLATPAK).studio.desktop application/x-roblox-rbxlx

tests:
	$(GO) test $(GOFLAGS) ./...

clean:
	rm -f vinegar robloxmutexer.exe

.PHONY: all install install-vinegar install-robloxmutexer install-desktop install-icons uninstall icons mime tests clean
