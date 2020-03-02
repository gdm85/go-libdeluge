all: test build bin/delugecli

build:
	go install

build-windows:
	GOOS=windows GOARCH=amd64 go install

test: *.go
	go test -v

bin/delugecli:
	go build -o $@ cli/cli.go

clean:
	rm -f bin/delugecli

.PHONY: all build test clean bin/delugecli build-windows
