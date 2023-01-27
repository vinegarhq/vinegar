build:
	go build -o bin/vinegar src/main.go

run:
	go run src/main.go

install:
	@echo "Nothing to install yet"
clean:
	rm bin/*

all:	build run
