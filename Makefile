VERSION = v1.5.9

PREFIX     = /usr
BINPREFIX  = $(PREFIX)/libexec/vinegar
APPPREFIX  = $(PREFIX)/share/applications
ICONPREFIX = $(PREFIX)/share/icons/hicolor

GO = go
GO_LDFLAGS = -s -w

VINEGAR_LDFLAGS = $(GO_LDFLAGS) -X main.BinPrefix=$(BINPREFIX) -X main.Version=$(VERSION)
VINEGAR_GOFLAGS = --tags nowayland,novulkan 

ROBLOX_ICONS = \
	icons/128/roblox-player.png icons/128/roblox-studio.png \
	icons/16/roblox-player.png icons/16/roblox-studio.png \
	icons/32/roblox-player.png icons/32/roblox-studio.png \
	icons/48/roblox-player.png icons/48/roblox-studio.png \
	icons/64/roblox-player.png icons/64/roblox-studio.png

VINEGAR_ICON = splash/vinegar.png

all: vinegar robloxmutexer.exe
icons: $(ROBLOX_ICONS) $(VINEGAR_ICON)
install: install-vinegar install-robloxmutexer install-desktop install-icons

vinegar:
	$(GO) build $(VINEGAR_GOFLAGS) $(GOFLAGS) -ldflags="$(VINEGAR_LDFLAGS)" ./cmd/vinegar

robloxmutexer.exe:
	GOOS=windows $(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/robloxmutexer

$(ROBLOX_ICONS): icons/roblox-player.svg icons/roblox-studio.svg
	rm -rf icons/16 icons/32 icons/48 icons/64 icons/128
	mkdir  icons/16 icons/32 icons/48 icons/64 icons/128
	convert -density 384 -background none $^ -resize 16x16   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -density 384 -background none $^ -resize 32x32   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -density 384 -background none $^ -resize 48x48   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -density 384 -background none $^ -resize 64x64   -set filename:f '%w/%t' 'icons/%[filename:f].png'
	convert -density 384 -background none $^ -resize 128x128 -set filename:f '%w/%t' 'icons/%[filename:f].png'
	
$(VINEGAR_ICON): icons/vinegar.svg
	# -fuzz 1% -trim +repage removes empty space, makes the image 44x64
	convert -density 384 -background none icons/vinegar.svg -resize 64x64 -fuzz 1% -trim +repage splash/vinegar.png

install-vinegar: vinegar
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar

install-robloxmutexer: robloxmutexer.exe
	install -Dm755 robloxmutexer.exe $(DESTDIR)$(BINPREFIX)/robloxmutexer.exe

install-desktop:
	mkdir -p $(DESTDIR)$(APPPREFIX)
	install -Dm644 desktop/vinegar.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop
	install -Dm644 desktop/roblox-app.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.app.desktop
	install -Dm644 desktop/roblox-player.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.player.desktop
	install -Dm644 desktop/roblox-studio.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.studio.desktop

install-icons: icons
	install -Dm644 icons/vinegar.svg $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.svg
	install -Dm644 icons/16/roblox-player.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/org.vinegarhq.Vinegar.player.png
	install -Dm644 icons/16/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/32/roblox-player.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/org.vinegarhq.Vinegar.player.png
	install -Dm644 icons/32/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/48/roblox-player.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/org.vinegarhq.Vinegar.player.png
	install -Dm644 icons/48/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/64/roblox-player.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/org.vinegarhq.Vinegar.player.png
	install -Dm644 icons/64/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 icons/128/roblox-player.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/org.vinegarhq.Vinegar.player.png
	install -Dm644 icons/128/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/org.vinegarhq.Vinegar.studio.png

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(BINPREFIX)/robloxmutexer.exe
	rm -f $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.player.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.studio.desktop
	rm -f $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.png
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/org.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/org.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/org.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/org.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/org.vinegarhq.Vinegar.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/org.vinegarhq.Vinegar.studio.png

	
mime:
	xdg-mime default org.vinegarhq.Vinegar.player.desktop x-scheme-handler/roblox-player
	xdg-mime default org.vinegarhq.Vinegar.player.desktop x-scheme-handler/roblox
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop x-scheme-handler/roblox-studio
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop x-scheme-handler/roblox-studio-auth
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop application/x-roblox-rbxl
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop application/x-roblox-rbxlx

tests:
	$(GO) test $(GOFLAGS) ./...

clean:
	rm -f vinegar robloxmutexer.exe

.PHONY: all install install-vinegar install-robloxmutexer install-desktop install-icons uninstall icons mime tests clean
