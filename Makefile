APP_NAME=payments-service

.PHONY: run test fmt build

run:
	go run ./cmd/api

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

build:
	go build -o bin/$(APP_NAME) ./cmd/api
	