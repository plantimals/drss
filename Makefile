
all: clean build run

.PHONY: clean build

build:
	go build -o ./drss main.go

dependencies:
	go mod download

run:
	./drss --storage `pwd`/feed

clean:
	rm drss; rm -rf feed
