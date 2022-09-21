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
	go test -coverprofile cover.out -v ./...

.PHONY: clean
clean:
	@rm -rf $(FILES)

### fmt_license: ensure license header is set on all files
fmt_license:
ifneq ($(shell command -v addlicense 2> /dev/null),)
	@echo 'addlicense -v -f license_header.txt **/*.go'
	@addlicense -v -f license_header.txt $$(find . -name '*.go')
else
	$(error "addlicense must be installed for this command: go install github.com/google/addlicense@latest")
endif

### check_fmt: Checks for missing licenses on files in repo
check_license:
  ifeq ($(shell command -v addlicense 2> /dev/null),)
	  $(error "error addlicense must be installed for this command: go install github.com/google/addlicense@latest")
  endif

	  if ! addlicense -check -f license_header.txt $$(find . -not -path '*/\.*' -name '*.go'); then \
	    echo "Licenses are not formatted; run 'make fmt_license'"; exit 1 ;\
	  fi \



### gosec - runs the gosec scanner for non-test files in this repo
.PHONY: gosec
gosec:
	# Run this command to install gosec, if not installed:
	# go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -no-fail -fmt=json -out=gosec.json -exclude-dir pkg/testingutil -exclude-dir tests ./...