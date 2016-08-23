all: test build bin/add-torrent

setup-gopath:
	mkdir -p .gopath
	if [ ! -L .gopath/src ]; then ln -s "$(CURDIR)/vendor" .gopath/src; fi
	if [ ! -L .gopath/src/github.com/gdm85/go-libdeluge ]; then ln -s "$(CURDIR)" .gopath/src/github.com/gdm85/go-libdeluge; fi

build: setup-gopath
	GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" go build

test: *.go
	GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" go test

bin/add-torrent: setup-gopath
	mkdir -p bin
	cd examples && GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" GOBIN="$(CURDIR)/bin/" go install add-torrent.go

clean:
	rm -f bin/add-torrent

.PHONY: all setup-gopath build test clean
