VERSION = v1.7.8

PREFIX     = /usr
APPPREFIX  = $(PREFIX)/share/applications
ICONPREFIX = $(PREFIX)/share/icons/hicolor

GO         = go
GO_LDFLAGS = -s -w -X main.Version=$(VERSION)

ROBLOX_ICONS = \
	assets/icons/128/roblox-studio.png \
	assets/icons/16/roblox-studio.png \
	assets/icons/32/roblox-studio.png \
	assets/icons/48/roblox-studio.png \
	assets/icons/64/roblox-studio.png

VINEGAR_ICON = splash/vinegar.png

all: vinegar
icons: $(ROBLOX_ICONS) $(VINEGAR_ICON)
install: install-vinegar install-desktop install-icons

vinegar:
	$(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/vinegar

$(ROBLOX_ICONS): assets/roblox-studio.svg
	rm -rf assets/icons
	mkdir -p assets/icons/128 assets/icons/64 assets/icons/48 assets/icons/32 assets/icons/16
	magick -density 384 -background none $^ -resize 16x16   assets/icons/16/roblox-studio.png
	magick -density 384 -background none $^ -resize 32x32   assets/icons/32/roblox-studio.png
	magick -density 384 -background none $^ -resize 48x48   assets/icons/48/roblox-studio.png
	magick -density 384 -background none $^ -resize 64x64   assets/icons/64/roblox-studio.png
	magick -density 384 -background none $^ -resize 128x128 assets/icons/128/roblox-studio.png
	
$(VINEGAR_ICON): assets/vinegar.svg
	# -fuzz 1% -trim +repage removes empty space, makes the image 44x64
	magick -density 384 -background none assets/vinegar.svg -resize 64x64 -fuzz 1% -trim +repage splash/vinegar.png

install-vinegar: vinegar assets/org.vinegarhq.Vinegar.metainfo.xml
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar
	install -Dm644 assets/org.vinegarhq.Vinegar.metainfo.xml -t $(DESTDIR)$(PREFIX)/share/metainfo

install-desktop:
	install -Dm644 assets/desktop/vinegar.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop
	install -Dm644 assets/desktop/roblox-app.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.app.desktop
	install -Dm644 assets/desktop/roblox-studio.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.studio.desktop

install-icons: icons
	install -Dm644 assets/vinegar.svg $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.svg
	install -Dm644 assets/icons/16/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/16x16/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 assets/icons/32/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/32x32/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 assets/icons/48/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/48x48/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 assets/icons/64/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/64x64/apps/org.vinegarhq.Vinegar.studio.png
	install -Dm644 assets/icons/128/roblox-studio.png $(DESTDIR)$(ICONPREFIX)/128x128/apps/org.vinegarhq.Vinegar.studio.png

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar
	rm -f $(DESTDIR)$(PREFIX)/share/metainfo/org.vinegarhq.Vinegar.metainfo.xml
	rm -f $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.app.desktop
	rm -f $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.studio.desktop
	rm -f $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.svg
	rm -f $(DESTDIR)$(ICONPREFIX)/16x16/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/32x32/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/48x48/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/64x64/apps/org.vinegarhq.Vinegar.studio.png
	rm -f $(DESTDIR)$(ICONPREFIX)/128x128/apps/org.vinegarhq.Vinegar.studio.png

mime:
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop x-scheme-handler/roblox-studio
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop x-scheme-handler/roblox-studio-auth
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop application/x-roblox-rbxl
	xdg-mime default org.vinegarhq.Vinegar.studio.desktop application/x-roblox-rbxlx

tests:
	$(GO) test $(GOFLAGS) ./...

clean:
	rm -f vinegar
	
.PHONY: all install install-vinegar install-desktop install-icons uninstall icons mime tests clean
