
all: clean build run

.PHONY: clean build

build:
	go build -o ./drsstools cmd/drsstools/drsstools.go

dependencies:
	go mod download

run:
	./drsstools --feedURL https://blog.golang.org/feed.atom

clean:
	rm drsstools
