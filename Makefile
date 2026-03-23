BINARY=terraform-provider-gestioip

.PHONY: build test fmt testacc testacc-3.2 testacc-3.5

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

fmt:
	gofmt -w .

testacc: testacc-3.2 testacc-3.5

testacc-3.2:
	TF_ACC=1 go test ./internal/provider -run TestAccGestioIP32Lifecycle -count=1 -v

testacc-3.5:
	TF_ACC=1 go test ./internal/provider -run TestAccGestioIP35Lifecycle -count=1 -v
