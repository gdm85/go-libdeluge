# go-libdeluge

Go library for native RPC connection to a [Deluge](http://deluge-torrent.org) daemon; it uses [go-rencode](https://github.com/gdm85/go-rencode/) for the RPC protocol serialization/deserialization.

[Release blog post](https://medium.com/where-do-we-go-now/accessing-a-deluge-server-with-go-d28a94e9b13f).

# License

[GNU GPL version 2](./LICENSE)

# Supported deluge versions

Both deluge v2.0+ and v1.3+ are supported; in order to use the modern deluge v2 daemon you must set `V2Daemon` to true in `delugeclient.Settings`.

# How to build

This project uses an automatically-provisioned GOPATH. Example init/building commands on a Linux system:

```
git submodule update --init --recursive
make
```

# How to use

The library by itself is a Go package and needs to be embedded in an UI or CLI application.

```go
	deluge := delugeclient.New(delugeclient.Settings{
		Hostname:              "localhost",
		Port:                  58846,
		Login:                 "localclient",
		Password:              "*************",
		V2Daemon:              v2daemon})

	// perform connection to Deluge server
	err := deluge.Connect()

	// ... use the 'deluge' client methods
```

To debug the library you may want to set `DebugSaveInteractions` to true.

## Example CLI application

An example CLI application is available through:
```
go get github.com/gdm85/go-libdeluge/delugecli
```

Example usage:

```sh
DELUGE_PASSWORD="mypassword" bin/delugecli -add magnet:?xt=urn:btih:C1939CA413B9AFCC34EA0CF3C128574E93FF6CB0&tr=http%3A%2F%2Ftorrent.ubuntu.com%3A6969%2Fannounce
```

This will start downloading the latest Ubuntu 14.04 LTS server ISO. Multiple magnet URIs are supported as command-line arguments; run `bin/delugecli` alone to see all available options and their description.
