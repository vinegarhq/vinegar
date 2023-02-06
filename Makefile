PREFIX     = /usr/local
APPPREFIX  = $(PREFIX)/share/applications
ICONPREFIX = $(PREFIX)/share/icons/hicolor

GOFLAGS = -ldflags="-s -w" -buildvcs=false

all: vinegar

vinegar:
	go build $(GOFLAGS) -o vinegar

install: vinegar
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar
	install -Dm644 desktop/app.desktop $(DESTDIR)$(APPPREFIX)/vinegar-app.desktop
	install -Dm644 desktop/player.desktop $(DESTDIR)$(APPPREFIX)/vinegar-roblox-player.desktop
	install -Dm644 desktop/studio.desktop $(DESTDIR)$(APPPREFIX)/vinegar-roblox-studio.desktop
	install -Dm644 icons/16/player.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/vinegar-roblox-player.png
	install -Dm644 icons/32/player.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/vinegar-roblox-player.png
	install -Dm644 icons/48/player.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/vinegar-roblox-player.png
	install -Dm644 icons/64/player.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/vinegar-roblox-player.png
	install -Dm644 icons/128/player.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/vinegar-roblox-player.png
	install -Dm644 icons/16/studio.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/vinegar-roblox-studio.png
	install -Dm644 icons/32/studio.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/vinegar-roblox-studio.png
	install -Dm644 icons/48/studio.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/vinegar-roblox-studio.png
	install -Dm644 icons/64/studio.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/vinegar-roblox-studio.png
	install -Dm644 icons/128/studio.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/vinegar-roblox-studio.png

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(APPPREFIX)/vinegar-roblox-app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/vinegar-roblox-player.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/vinegar-roblox-studio.desktop
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/vinegar-roblox-player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/vinegar-roblox-player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/vinegar-roblox-player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/vinegar-roblox-player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/vinegar-roblox-player.png
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/vinegar-roblox-studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/vinegar-roblox-studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/vinegar-roblox-studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/vinegar-roblox-studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/vinegar-roblox-studio.png

clean:
	rm -f vinegar

.PHONY: all vinegar clean install uninstall
