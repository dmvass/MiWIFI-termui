NAME=miwifi-termui
WRKDIR=$(pwd)
VERSION=$(shell git describe --tags || echo n/a)

all: fmt build

fmt:
	go fmt ./...

clean:
	rm -rf $(WRKDIR)/build/

build: clean
	go build -ldflags "-s -X main.version=$(VERSION)" -o bin/$(NAME) ./cmd/$(NAME)

lint:
	golangci-lint run

test:
	go test ./...
