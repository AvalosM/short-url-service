all: binshorturl
.PHONY: all

binshorturl:
	 go build -o binshorturl ./cmd/shorturl/main.go

clean:
	 rm -f binshorturl
.PHONY: clean

lint:
	 golangci-lint run --config golangci.yml ${GCI_FLAGS}
.PHONY: lint

lint-fix:
	 make lint GCI_FLAGS="--fix"
.PHONY: lint-fix

test:
	 go test -v ./...

docs:
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/shorturl/main.go --parseDepth 1 --output ./docs/swagger
.PHONY: docs

generate:
	 go install go.uber.org/mock/mockgen@latest
	 go generate ./...
.PHONY: generate