package ?= default

build:
	go build -o bin/${package} ./...

test:
	go test ./...

run:
	go run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...


