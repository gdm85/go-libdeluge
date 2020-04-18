all: bin/delugecli test

bin/delugecli-windows:
	GOOS=windows GOARCH=amd64 go build -o $@ ./delugecli

test: *.go
	go test -v

bin/delugecli:
	go build -o $@ ./delugecli

clean:
	rm -f bin/delugecli bin/delugecli-windows

.PHONY: all build test clean bin/delugecli bin/delugecli-windows
