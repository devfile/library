FILES := main

default: main

.PHONY: all
all:  deps main test

.PHONY: deps
deps:
	 go mod tidy

.PHONY: test
test:
	 go test ./...

.PHONY: main
main:
	 go build main.go

.PHONY: clean
clean:
	@rm -rf $(FILES)

