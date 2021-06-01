FILES := main

default: bin

.PHONY: all
all:  gomod_tidy gofmt bin test

.PHONY: gomod_tidy
gomod_tidy:
	 go mod tidy

.PHONY: gofmt
gofmt:
	go fmt -x ./...

.PHONY: bin
bin:
	 go build *.go

.PHONY: test
test:
	go test -coverprofile tests/v2/lib-test-coverage.out -v ./...
	go tool cover -html=tests/v2/lib-test-coverage.out -o tests/v2/lib-test-coverage.html

.PHONY: clean
clean:
	@rm -rf $(FILES)

