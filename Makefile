FILES := main

default: bin

.PHONY: all
all:  gomod_tidy gofmt2 gofmt bin test

.PHONY: gomod_tidy
gomod_tidy:
	 go mod tidy

.PHONY: gofmt
gofmt:
	go fmt -x ./...

.PHONY: gofmt2
gofmt2:
	gofmt -d -e ./pkg/tests/parserv200/*.go

.PHONY: bin
bin:
	 go build main.go

.PHONY: test
test:
	 go test ./...

.PHONY: clean
clean:
	@rm -rf $(FILES)

