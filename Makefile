
all: clean build run

.PHONY: clean build

build:
	go build -o ./drss main.go

dependencies:
	go mod download

run:
	./drss --feedURL https://blog.golang.org/feed.atom

clean:
	rm drss; rm -rf feed
