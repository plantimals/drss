
all: build run

.PHONY: clean build

build:
	go build -o ./drss main.go

dependencies:
	go mod download

run:
	./drss episode.json

clean:
	rm drss
