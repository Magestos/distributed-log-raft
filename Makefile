package ?= default

build:
	go build -o bin/${package} ./...

test:
	go test ./...

run:
	go run ./cmd/node

fmt:
	gofmt -w .

vet:
	go vet ./...

