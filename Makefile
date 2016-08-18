build-all: build bin/add-torrent

build:
	mkdir -p .gopath
	if [ ! -L .gopath/src ]; then ln -s "$(CURDIR)/vendor" .gopath/src; fi
	if [ ! -L .gopath/src/github.com/gdm85/go-libdeluge ]; then ln -s "$(CURDIR)" .gopath/src/github.com/gdm85/go-libdeluge; fi
	GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" go build

bin/add-torrent: build
	mkdir -p bin
	cd examples && GO15VENDOREXPERIMENT=1 GOPATH="$(CURDIR)/.gopath" GOBIN="$(CURDIR)/bin/" go install add-torrent.go

clean:
	rm -f bin/add-torrent

.PHONY: build-all build clean
