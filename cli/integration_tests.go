package main

import (
	"errors"
	"time"

	delugeclient "github.com/gdm85/go-libdeluge"
)

const (
	testMagnetHash = "C1939CA413B9AFCC34EA0CF3C128574E93FF6CB0"
	testMagnetURI  = `magnet:?xt=urn:btih:C1939CA413B9AFCC34EA0CF3C128574E93FF6CB0&tr=http%3A%2F%2Ftorrent.ubuntu.com%3A6969%2Fannounce`
)

func runAllIntegrationTests(deluge *delugeclient.Client) error {
	_, err := deluge.DaemonVersion()
	if err != nil {
		return err
	}

	methods, err := deluge.MethodsList()
	if err != nil {
		return err
	}
	if len(methods) == 0 {
		return errors.New("no methods returned")
	}

	_, err = deluge.GetFreeSpace("")
	if err != nil {
		return err
	}

	_, err = deluge.GetAvailablePlugins()
	if err != nil {
		return err
	}

	_, err = deluge.GetEnabledPlugins()
	if err != nil {
		return err
	}

	if v2daemon {
		_, err = deluge.KnownAccounts()
		if err != nil {
			return err
		}
	}

	torrentHash, err := deluge.AddTorrentMagnet(testMagnetURI, nil)
	if err != nil {
		return err
	}
	if torrentHash == "" {
		return errors.New("torrent was not added")
	}

	const maxAttempts = 10
	found := false
	attempts := 0
	for !found {
		torrents, err := deluge.TorrentsStatus(delugeclient.StateUnspecified, nil)
		if err != nil {
			return err
		}

		for id := range torrents {
			if id == testMagnetHash {
				found = true
				break
			}
		}
		if !found {
			attempts++
			if attempts == maxAttempts {
				return errors.New("cannot find torrent")
			}
			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}
