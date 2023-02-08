PREFIX     = /usr/local
APPPREFIX  = $(PREFIX)/share/applications
ICONPREFIX = $(PREFIX)/share/icons/hicolor

GOFLAGS = -ldflags="-s -w" -buildvcs=false
GO = go

all: vinegar

vinegar:
	$(GO) build $(GOFLAGS) -o vinegar

install: vinegar
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar
	install -Dm644 desktop/com.github.vinegar.app.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.app.desktop
	install -Dm644 desktop/com.github.vinegar.player.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.player.desktop
	install -Dm644 desktop/com.github.vinegar.studio.desktop $(DESTDIR)$(APPPREFIX)/com.github.vinegar.studio.desktop
	install -Dm644 icons/16/com.github.vinegar.roblox.player.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/com.github.vinegar.roblox.player.png
	install -Dm644 icons/32/com.github.vinegar.roblox.player.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/com.github.vinegar.roblox.player.png
	install -Dm644 icons/48/com.github.vinegar.roblox.player.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/com.github.vinegar.roblox.player.png
	install -Dm644 icons/64/com.github.vinegar.roblox.player.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/com.github.vinegar.roblox.player.png
	install -Dm644 icons/128/com.github.vinegar.roblox.player.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/com.github.vinegar.roblox.player.png
	install -Dm644 icons/16/com.github.vinegar.roblox.studio.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/com.github.vinegar.roblox.studio.png
	install -Dm644 icons/32/com.github.vinegar.roblox.studio.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/com.github.vinegar.roblox.studio.png
	install -Dm644 icons/48/com.github.vinegar.roblox.studio.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/com.github.vinegar.roblox.studio.png
	install -Dm644 icons/64/com.github.vinegar.roblox.studio.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/com.github.vinegar.roblox.studio.png
	install -Dm644 icons/128/com.github.vinegar.roblox.studio.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/com.github.vinegar.roblox.studio.png

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.player.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/com.github.vinegar.studio.desktop
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/com.github.vinegar.roblox.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/com.github.vinegar.roblox.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/com.github.vinegar.roblox.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/com.github.vinegar.roblox.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/com.github.vinegar.roblox.player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/com.github.vinegar.roblox.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/com.github.vinegar.roblox.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/com.github.vinegar.roblox.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/com.github.vinegar.roblox.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/com.github.vinegar.roblox.studio.png

clean:
	rm -f vinegar

.PHONY: all vinegar clean install uninstall
