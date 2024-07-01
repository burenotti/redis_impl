BINARY_NAME = redis
COVERAGE_VAR = COVERAGE

all: clean build test

_init:
	@[[ ! -d build ]] && mkdir build || true

build: _init
	GOOS=darwin GOARCH=arm64 go build -o build/$(BINARY_NAME)_darwin_arm64 -trimpath -ldflags="-s -w" ./cmd/redis/
	GOOS=linux GOARCH=amd64 go build -o build/$(BINARY_NAME)_linux_amd64 -trimpath -ldflags="-s -w" ./cmd/redis/
	GOOS=windows GOARCH=amd64 go build -o build/$(BINARY_NAME)_windows_amd64.exe -trimpath -ldflags="-s -w" ./cmd/redis/

test: _init
	go test -race -coverprofile ./build/coverage.out ./...
	go tool cover -var $(COVERAGE_VAR) -html ./build/coverage.out -o ./build/coverage.html ; \
	echo "$$$(COVERAGE_VAR)"

clean:
	rm -rf ./build/

run:
	go run ./cmd/redis -config ./config/config.dist.yaml