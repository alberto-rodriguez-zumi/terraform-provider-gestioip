BINARY=terraform-provider-gestioip

.PHONY: build test fmt

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

fmt:
	gofmt -w .

