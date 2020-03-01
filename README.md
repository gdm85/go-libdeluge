# go-libdeluge

Go library for native RPC connection to a [Deluge](http://deluge-torrent.org) daemon; it uses [go-rencode](https://github.com/gdm85/go-rencode/) for the RPC protocol serialization/deserialization.

[Release blog post](https://medium.com/where-do-we-go-now/accessing-a-deluge-server-with-go-d28a94e9b13f).

# License

[GNU GPL version 2](./LICENSE)

# How to build

This project uses Go modules. You can build it with `make`:

```
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

# Supported deluge versions

Both deluge v2.0+ and v1.3+ are supported; in order to use the modern deluge v2 daemon you must set `V2Daemon` to true in `delugeclient.Settings`.

## RPC API supported methods

* [x] `daemon.login`
* [x] `daemon.info`
* [ ] `daemon.authorized_call`
* [x] `daemon.get_method_list`
* [ ] `daemon.get_version`
* [ ] `daemon.shutdown`
* [ ] `core.add_torrent_file`
* [ ] `core.add_torrent_file_async`
* [ ] `core.add_torrent_files`
* [x] `core.add_torrent_magnet`
* [x] `core.add_torrent_url`
* [ ] `core.connect_peer`
* [x] `core.create_account`
* [ ] `core.create_torrent`
* [ ] `core.disable_plugin`
* [ ] `core.enable_plugin`
* [ ] `core.force_reannounce`
* [ ] `core.force_recheck`
* [ ] `core.get_auth_levels_mappings`
* [ ] `core.get_available_plugins`
* [ ] `core.get_completion_paths`
* [ ] `core.get_config`
* [ ] `core.get_config_value`
* [ ] `core.get_config_values`
* [x] `core.get_enabled_plugins`
* [ ] `core.get_external_ip`
* [ ] `core.get_filter_tree`
* [x] `core.get_free_space`
* [x] `core.get_known_accounts`
* [ ] `core.get_libtorrent_version`
* [ ] `core.get_listen_port`
* [ ] `core.get_path_size`
* [ ] `core.get_proxy`
* [x] `core.get_session_state`
* [ ] `core.get_session_status`
* [x] `core.get_torrent_status`
* [x] `core.get_torrents_status`
* [ ] `core.glob`
* [ ] `core.is_session_paused`
* [x] `core.move_storage`
* [ ] `core.pause_session`
* [x] `core.pause_torrent`
* [x] `core.pause_torrents`
* [ ] `core.prefetch_magnet_metadata`
* [ ] `core.queue_bottom`
* [ ] `core.queue_down`
* [ ] `core.queue_top`
* [ ] `core.queue_up`
* [x] `core.remove_account`
* [x] `core.remove_torrent`
* [x] `core.remove_torrents`
* [ ] `core.rename_files`
* [ ] `core.rename_folder`
* [ ] `core.rescan_plugins`
* [ ] `core.resume_session`
* [x] `core.resume_torrent`
* [x] `core.resume_torrents`
* [ ] `core.set_config`
* [x] `core.set_torrent_options`
* [x] `core.set_torrent_trackers`
* [ ] `core.test_listen_port`
* [x] `core.update_account`
* [ ] `core.upload_plugin`
