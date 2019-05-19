all: test build bin/delugecli

build:
	go build

test: *.go
	go test -v

bin/delugecli:
	mkdir -p bin
	GOBIN="$(CURDIR)/bin/" go install cli/cli.go
	mv bin/cli bin/delugecli

clean:
	rm -f bin/delugecli

.PHONY: all build test clean bin/delugecli
