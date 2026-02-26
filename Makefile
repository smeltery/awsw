.PHONY: build test lint clean install

VERSION ?= dev

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o awsw .

test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

lint:
	golangci-lint run

clean:
	rm -f awsw coverage.out

install: build
	cp awsw $(GOPATH)/bin/awsw 2>/dev/null || cp awsw /usr/local/bin/awsw
