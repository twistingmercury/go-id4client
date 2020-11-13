default: build

build:
	go build ./...

test:
	go clean -testcache ./...
	go test -coverprofile=coverage.out ./...

cover:
	go tool cover -html=coverage.out