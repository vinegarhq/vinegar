.POSIX:

VERSION = v1.7.8

PREFIX      ?= /usr
DATAPREFIX  = $(PREFIX)/share/vinegar
APPPREFIX   = $(PREFIX)/share/applications
ICONPREFIX  = $(PREFIX)/share/icons/hicolor
LIBPREFIX   = $(PREFIX)/lib
LAYERPREFIX = $(PREFIX)/share/vulkan/explicit_layer.d

CXX ?= c++

GO         ?= go
GO_LDFLAGS ?= -s -w -X main.version=$(VERSION)

# for automatically re-building vinegar
SOURCES != find . -type f -name "*.go"

all: vinegar layer/libVkLayer_VINEGAR_VinegarLayer.so

vinegar: $(SOURCES) cmd/vinegar/vinegar.gresource
	$(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/vinegar

cmd/vinegar/vinegar.gresource: data/vinegar.gresource.xml data/ui/vinegar.cmb
	glib-compile-resources --sourcedir=data --target=cmd/vinegar/vinegar.gresource data/vinegar.gresource.xml

layer/libVkLayer_VINEGAR_VinegarLayer.so: layer/vinegar_layer.cpp
	$(CXX) -shared -fPIC $< -o $@

install: vinegar
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar
	install -Dm644 data/org.vinegarhq.Vinegar.metainfo.xml $(DESTDIR)$(PREFIX)/share/metainfo/
	install -Dm644 data/desktop/vinegar.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop
	install -Dm644 data/desktop/roblox-studio.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.studio.desktop
	install -Dm644 data/icons/vinegar.svg $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.svg
	install -Dm644 data/icons/roblox-studio.svg $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.studio.svg
	install -Dm644 layer/libVkLayer_VINEGAR_VinegarLayer.so $(DESTDIR)$(LIBPREFIX)/
	install -Dm644 layer/VkLayer_VINEGAR_VinegarLayer.json $(DESTDIR)$(LAYERPREFIX)/
	gtk-update-icon-cache -f -t $(DESTDIR)$(ICONPREFIX) ||:

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar \
		$(DESTDIR)$(PREFIX)/share/metainfo/org.vinegarhq.Vinegar.metainfo.xml \
		$(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop \
		$(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.studio.desktop \
		$(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.svg \
		$(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.studio.svg \
		$(DESTDIR)$(LIBPREFIX)/libVkLayer_VINEGAR_VinegarLayer.so \
		$(DESTDIR)$(LAYERPREFIX)/VkLayer_VINEGAR_VinegarLayer.json

clean:
	rm -f vinegar layer/libVkLayer_VINEGAR_VinegarLayer.so
	
.PHONY: all install uninstall clean
