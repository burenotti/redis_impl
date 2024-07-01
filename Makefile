BINARY_NAME = redis
COVERAGE_VAR = COVERAGE


BUILD_TAGS = "integration,unit"
unit-tests: BUILD_TAGS = "unit"
all-tests: BUILD_TAGS = "disabled,integration,unit"

all: clean build test

run:
	go run ./cmd/redis -config ./config/config.dist.yaml


build: _init
	GOOS=darwin GOARCH=arm64 go build -o build/$(BINARY_NAME)_darwin_arm64 -trimpath -ldflags="-s -w" ./cmd/redis/
	GOOS=linux GOARCH=amd64 go build -o build/$(BINARY_NAME)_linux_amd64 -trimpath -ldflags="-s -w" ./cmd/redis/
	GOOS=windows GOARCH=amd64 go build -o build/$(BINARY_NAME)_windows_amd64.exe -trimpath -ldflags="-s -w" ./cmd/redis/

test: _init
	go test -tags=$(BUILD_TAGS) -race -coverprofile ./build/coverage.out ./...
	go tool cover -html ./build/coverage.out -o ./build/coverage.html


unit-tests: test


all-tests: test


clean:
	rm -rf ./build/


_init:
	@test ! -d build && mkdir build || true
