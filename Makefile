build:
	go build -ldflags="-s -w" -o bin/vinegar src/main.go

run:
	go run src/main.go

uninstall:
	rm ~/.local/bin/vinegar
	rm ~/.local/share/applications/com.github.vinegar.app.desktop
	rm ~/.local/share/icons/hicolor/128x128/apps/com.github.vinegar.png

install:
	mkdir -p ~/.local/bin
	install -Dm00755 bin/vinegar ~/.local/bin/
	mkdir -p ~/.local/share/applications
	install -Dm00644 com.github.vinegar.app.desktop ~/.local/share/applications/
	install -Dm00644 com.github.vinegar.player.desktop ~/.local/share/applications/
	install -Dm00644 com.github.vinegar.studio.desktop ~/.local/share/applications/
	mkdir -p ~/.local/share/icons/hicolor/128x128/apps
	install -Dm00644 com.github.vinegar.png ~/.local/share/icons/hicolor/128x128/apps/

clean:
	rm bin/*

all:	build install

# Someone make a better makefile with install instead of cp, please.
