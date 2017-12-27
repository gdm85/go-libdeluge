all: test build bin/delugecli

setup-gopath:
	mkdir -p .gopath
	if [ ! -L .gopath/src ]; then ln -s "$(CURDIR)/vendor" .gopath/src; fi
	if [ ! -L .gopath/src/github.com/gdm85/go-libdeluge ]; then ln -s "$(CURDIR)" .gopath/src/github.com/gdm85/go-libdeluge; fi

build: setup-gopath
	GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" go build

test: *.go
	GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" go test -v

bin/delugecli: setup-gopath
	mkdir -p bin
	GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" GOBIN="$(CURDIR)/bin/" go install cli/cli.go
	mv bin/cli bin/delugecli

clean:
	rm -f bin/delugecli

.PHONY: all setup-gopath build test clean bin/delugecli
