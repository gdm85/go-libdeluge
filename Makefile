all: test build bin/delugecli

build:
	go build

test: *.go
	go test -v

bin/delugecli:
	go build -o $@ cli/cli.go

clean:
	rm -f bin/delugecli

.PHONY: all build test clean bin/delugecli
