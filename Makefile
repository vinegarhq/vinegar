build:
	go build -ldflags="-s -w" -o bin/vinegar src/main.go

run:
	go run src/main.go

install:
	@echo "Nothing to install yet"
	bin/vinegar
clean:
	rm bin/*

all:	build run
