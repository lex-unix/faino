MAIN_PACKAGE_PATH := ./cmd/faino
BINARY_NAME := faino

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## build: build the application
.PHONY: build
build:
	go build -o=./bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/integration: run all integration tests
.PHONY: test/integration
test/integration:
	go test -v -tags=integration -count=1  ./test/integration/...

## test/compose/down: stop and remove docker compose containers
.PHONY: test/compose/down
test/compose/down:
	@docker compose -f ./test/integration/docker-compose.yml down -t 1

## test/compose/start: stop and remove docker compose containers
.PHONY: test/compose/start
test/compose/start:
	@docker compose -f ./test/integration/docker-compose.yml up --build --detach
	@docker compose -f ./test/integration/docker-compose.yml exec --workdir / deployer  ./setup.sh
