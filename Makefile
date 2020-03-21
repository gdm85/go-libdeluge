all: bin/delugecli test

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/delugecli-windows ./cli

test: *.go
	go test -v

bin/delugecli:
	go build -o $@ ./cli

clean:
	rm -f bin/delugecli

.PHONY: all build test clean bin/delugecli build-windows
