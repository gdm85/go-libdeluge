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
	// you can use NewV1 to create a client for Deluge v1.3
	deluge := delugeclient.NewV2(delugeclient.Settings{
		Hostname:              "localhost",
		Port:                  58846,
		Login:                 "localclient",
		Password:              "*************",
	})

	// perform connection to Deluge server
	err := deluge.Connect()

	// ... use the client methods
```

To debug the library you may want to set `DebugServerResponses` to true.

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

Both deluge v2.0+ and v1.3+ are supported with the two different constructors `NewV2` and `NewV1`.

# RPC API supported methods

* [x] `daemon.login`
* [x] `daemon.info`
* [ ] `daemon.authorized_call`
* [x] `daemon.get_method_list`
* [ ] `daemon.get_version`
* [ ] `daemon.shutdown`
* [x] `core.add_torrent_file`
* [ ] `core.add_torrent_file_async`
* [ ] `core.add_torrent_files`
* [x] `core.add_torrent_magnet`
* [x] `core.add_torrent_url`
* [ ] `core.connect_peer`
* [x] `core.create_account`
* [ ] `core.create_torrent`
* [x] `core.disable_plugin`
* [x] `core.enable_plugin`
* [x] `core.force_reannounce`
* [ ] `core.force_recheck`
* [ ] `core.get_auth_levels_mappings`
* [x] `core.get_available_plugins`
* [ ] `core.get_completion_paths`
* [ ] `core.get_config`
* [ ] `core.get_config_value`
* [ ] `core.get_config_values`
* [x] `core.get_enabled_plugins`
* [ ] `core.get_external_ip`
* [ ] `core.get_filter_tree`
* [x] `core.get_free_space`
* [x] `core.get_known_accounts`
* [x] `core.get_libtorrent_version`
* [x] `core.get_listen_port`
* [ ] `core.get_path_size`
* [ ] `core.get_proxy`
* [x] `core.get_session_state`
* [x] `core.get_session_status`
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
* [x] `core.test_listen_port`
* [x] `core.update_account`
* [ ] `core.upload_plugin`

# Plugins

Plugins can be used by calling the relative method and checking if the result is not nil, example:

```go
	p, err := deluge.LabelPlugin()
	if err != nil {
		panic(err)
	}
	if p == nil {
		println("Label plugin not available")
		return
	}

	// call plugin methods
	labelsByTorrent, err := p.GetTorrentsLabels(delugeclient.StateUnspecified, nil)
```

## Label

### RPC API supported methods

* [x] `label.add`
* [ ] `label.get_config`
* [x] `label.get_labels`
* [ ] `label.get_options`
* [x] `label.remove`
* [ ] `label.set_config`
* [ ] `label.set_options`
* [x] `label.set_torrent`
