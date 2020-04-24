NAME=miwifi-cli
WRKDIR=$(pwd)

all: fmt build

fmt:
	go fmt ./...

clean:
	rm -rf $(WRKDIR)/build/

build: clean
	go build -o bin/$(NAME) ./cmd/$(NAME)
