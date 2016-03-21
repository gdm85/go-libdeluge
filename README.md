# go-libdeluge

Go library for native RPC connection to a [Deluge](http://deluge-torrent.org) daemon; it uses [go-rencode](https://github.com/gdm85/go-rencode/) for the RPC protocol serialization/deserialization.

[Release blog post](https://medium.com/where-do-we-go-now/accessing-a-deluge-server-with-go-d28a94e9b13f).

# License

[GNU GPL version 2](./LICENSE)

# How to build

This project uses an automatically-provisioned GOPATH. Example init/building commands on a Linux system:

```
git submodule update --init --recursive
make
```

# How to use

The library by itself is a Go package and needs to be embedded in an UI or CLI application.

## Example: add-torrent
The example add-torrent CLI program provided under [examples](https://github.com/gdm85/go-libdeluge/tree/master/examples) can be used to add a torrent magnet URI for download:

```
DELUGE_PASSWORD="mypassword" bin/add-torrent magnet:?xt=urn:btih:C1939CA413B9AFCC34EA0CF3C128574E93FF6CB0&tr=http%3A%2F%2Ftorrent.ubuntu.com%3A6969%2Fannounce
```

This will start downloading the latest Ubuntu 14.04 LTS server ISO. Multiple magnet URIs are supported as command-line arguments; run `bin/add-torrent` alone to see all available options and their description.
