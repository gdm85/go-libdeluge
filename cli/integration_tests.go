package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	delugeclient "github.com/gdm85/go-libdeluge"
)

const (
	testMagnetHash = "c1939ca413b9afcc34ea0cf3c128574e93ff6cb0"
	testMagnetURI  = `magnet:?xt=urn:btih:c1939ca413b9afcc34ea0cf3c128574e93ff6cb0&tr=http%3A%2F%2Ftorrent.ubuntu.com%3A6969%2Fannounce`
)

func runAllIntegrationTests(settings delugeclient.Settings) error {
	var deluge delugeclient.DelugeClient
	var c *delugeclient.Client
	settings.DebugServerResponses = true

	if v2daemon {
		cli := delugeclient.NewV2(settings)
		c = &cli.Client
		deluge = cli
	} else {
		c = delugeclient.NewV1(settings)
		deluge = c
	}

	// perform connection to Deluge server
	err := deluge.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
		os.Exit(3)
	}
	defer deluge.Close()

	_, err = deluge.DaemonVersion()
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
		deluge := deluge.(delugeclient.DelugeClientV2)
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

	const maxAttempts = 1
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

func printServerResponse(c *delugeclient.Client) {
	if len(c.DebugServerResponses) != 1 {
		panic("BUG: expected exactly one response")
	}

	// store response for testing/development
	buf := c.DebugServerResponses[0]
	fmt.Println("last call received contained", buf.Len(), "compressed bytes")

	fmt.Printf("payload: %X\n", buf.Bytes())

	c.DebugServerResponses = nil

}
