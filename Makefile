build:
	go build -ldflags="-s -w" -o bin/vinegar src/main.go

run:
	go run src/main.go

uninstall:
	rm ~/.local/bin/vinegar
	rm ~/.local/share/applications/com.github.vinegar.app.desktop
	rm ~/.local/share/icons/hicolor/128x128/apps/com.github.vinegar.png

install:
	cp bin/vinegar ~/.local/bin
	cp com.github.vinegar.app.desktop ~/.local/share/applications
	cp com.github.vinegar.png ~/.local/share/icons/hicolor/128x128/apps

clean:
	rm bin/*

all:	build install

# Someone make a better makefile with install instead of cp, please.
