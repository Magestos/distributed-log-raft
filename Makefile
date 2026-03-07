package ?= default
CONFIG ?= internal/config/config.yml
RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
RUN_CONFIG := $(or $(word 2,$(MAKECMDGOALS)),$(CONFIG))

ifneq ($(RUN_ARGS),)
.PHONY: $(RUN_ARGS)
$(eval $(RUN_ARGS):;@:)
endif

.PHONY: build test run fmt vet

build:
	go build -o bin/${package} ./...

test:
	go test ./...

run:
	go run ./cmd/node -config $(RUN_CONFIG)

fmt:
	gofmt -w .

vet:
	go vet ./...
