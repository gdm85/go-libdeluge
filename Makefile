all: bin/delugecli test

bin/delugecli:
	go build -o $@ ./delugecli

bin/delugecli-windows:
	GOOS=windows GOARCH=amd64 go build -o $@ ./delugecli

test: *.go
	go test -v

integration:
	go test -v -tags=integration,integration_v1 -c ./integration -o bin/inttest1
	go test -v -tags=integration,integration_v2 -c ./integration -o bin/inttest2

clean:
	rm -f bin/delugecli bin/delugecli-windows

.PHONY: all build test clean bin/delugecli bin/delugecli-windows integration
