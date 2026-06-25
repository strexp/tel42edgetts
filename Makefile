.PHONY: build clean check

APP_NAME = tel42edgetts
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

LDFLAGS = -s -w -extldflags "-static" -X main.Version=$(VERSION)

build:
	go mod tidy
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(APP_NAME) ./cmd/tel42edgetts

check: build
	go test ./... -v

clean:
	rm -f $(APP_NAME)
