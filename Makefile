.POSIX:

VERSION = v1.8.1

PREFIX      ?= /usr
DATAPREFIX  = $(PREFIX)/share/vinegar
APPPREFIX   = $(PREFIX)/share/applications
MIMEPREFIX  = $(PREFIX)/share/mime
ICONPREFIX  = $(PREFIX)/share/icons/hicolor
LIBPREFIX   = $(PREFIX)/lib
LAYERPREFIX = $(PREFIX)/share/vulkan/explicit_layer.d

CXX ?= c++

GO         ?= go
GO_LDFLAGS ?= -s -w -X main.version=$(VERSION)

SOURCES != find . -type f -name "*.go" # for automatically re-building vinegar

RESOURCE = internal/gtkutil/vinegar.gresource

all: vinegar layer/libVkLayer_VINEGAR_VinegarLayer.so

vinegar: $(SOURCES) $(RESOURCE)
	$(GO) build $(GOFLAGS) -ldflags="$(GO_LDFLAGS)" ./cmd/vinegar

$(RESOURCE): data/vinegar.gresource.xml data/ui/vinegar.cmb
	glib-compile-resources --sourcedir=data --target=$(RESOURCE) data/vinegar.gresource.xml

layer/libVkLayer_VINEGAR_VinegarLayer.so:
	$(CXX) -shared -fPIC `pkg-config --cflags vulkan` layer/vinegar_layer.cpp -o $@

install: all
	install -Dm755 vinegar $(DESTDIR)$(PREFIX)/bin/vinegar
	install -Dm644 data/org.vinegarhq.Vinegar.metainfo.xml -t $(DESTDIR)$(PREFIX)/share/metainfo
	install -Dm644 data/desktop/vinegar.desktop $(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop
	install -Dm644 data/vinegar-mime.xml -t $(DESTDIR)$(MIMEPREFIX)/packages
	install -Dm644 data/icons/vinegar.svg $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.svg
	install -Dm644 data/icons/roblox-studio.svg $(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.studio.svg
	install -Dm644 layer/libVkLayer_VINEGAR_VinegarLayer.so $(DESTDIR)$(LIBPREFIX)/libVkLayer_VINEGAR_VinegarLayer.so
	install -Dm644 layer/VkLayer_VINEGAR_VinegarLayer.json $(DESTDIR)$(LAYERPREFIX)/VkLayer_VINEGAR_VinegarLayer.json
	gtk-update-icon-cache $(DESTDIR)$(ICONPREFIX) ||:
	update-desktop-database $(DESTDIR)$(APPPREFIX) ||:
	update-mime-database $(DESTDIR)$(MIMEPREFIX) ||:

uninstall:
	# Retain removal of old studio desktop & icon files
	rm -f $(DESTDIR)$(PREFIX)/bin/vinegar \
		$(DESTDIR)$(PREFIX)/share/metainfo/org.vinegarhq.Vinegar.metainfo.xml \
		$(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.desktop \
		$(DESTDIR)$(APPPREFIX)/org.vinegarhq.Vinegar.studio.desktop \
		$(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.svg \
		$(DESTDIR)$(ICONPREFIX)/scalable/apps/org.vinegarhq.Vinegar.studio.svg \
		$(DESTDIR)$(LIBPREFIX)/libVkLayer_VINEGAR_VinegarLayer.so \
		$(DESTDIR)$(LAYERPREFIX)/VkLayer_VINEGAR_VinegarLayer.json

clean:
	rm -f vinegar layer/libVkLayer_VINEGAR_VinegarLayer.so $(RESOURCE)
	
.PHONY: all install uninstall clean
