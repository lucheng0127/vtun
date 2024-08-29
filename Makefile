.DEFAULT_GOAL := build

.PHONY: dir
dir:
	mkdir -p build/amd64 build/arm64

.PHONY: build-amd64
build-amd64:
	cp wintun/amd64/wintun.dll pkg/utils
	GOOS=linux GOARCH=amd64 go build -o build/amd64/vtun main.go
	GOOS=darwin GOARCH=amd64 go build -o build/amd64/vtun.darwin main.go
	GOOS=windows GOARCH=amd64 go build -o build/amd64/vtun.exe main.go

.PHONY: build-arm64
build-arm64:
	cp wintun/arm64/wintun.dll pkg/utils
	GOOS=linux GOARCH=arm64 go build -o build/arm64/vtun main.go
	GOOS=darwin GOARCH=arm64 go build -o build/arm64/vtun.darwin main.go
	GOOS=windows GOARCH=arm64 go build -o build/arm64/vtun.exe main.go

.PHONY: build
build: dir build-amd64 build-arm64

.PHONY: new-key
new-key:
	sh scripts/generate_key.sh

.PHONY: clean
clean:
	rm -f pkg/utils/wintun.dll
	rm -rf build
	rm -f index_*.html

.PHONY: test
test:
	go test -v -gcflags=-l ./...
