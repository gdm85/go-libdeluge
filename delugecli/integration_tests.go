// +build integration

// go-libdeluge v0.5.1 - a native deluge RPC client library
// Copyright (C) 2015~2020 gdm85 - https://github.com/gdm85/go-libdeluge/
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

// Command line util to issue requests against a headless deluge server.

package main

import (
	"errors"
	"fmt"
	"os"

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
	printServerResponse("DaemonLogin", c)

	_, err = deluge.DaemonVersion()
	if err != nil {
		return err
	}
	printServerResponse("DaemonVersion", c)

	methods, err := deluge.MethodsList()
	if err != nil {
		return err
	}
	printServerResponse("MethodsList", c)
	if len(methods) == 0 {
		return errors.New("no methods returned")
	}

	_, err = deluge.GetFreeSpace("")
	if err != nil {
		return err
	}
	printServerResponse("GetFreeSpace", c)

	_, err = deluge.GetAvailablePlugins()
	if err != nil {
		return err
	}
	printServerResponse("GetAvailablePlugins", c)

	_, err = deluge.GetEnabledPlugins()
	if err != nil {
		return err
	}
	printServerResponse("GetEnabledPlugins", c)

	if v2daemon {
		deluge := deluge.(delugeclient.V2)
		_, err = deluge.KnownAccounts()
		if err != nil {
			return err
		}
		printServerResponse("KnownAccounts", c)
	}

	torrentHash, err := deluge.AddTorrentMagnet(testMagnetURI, nil)
	if err != nil {
		return err
	}
	printServerResponse("AddTorrentMagnet", c)
	if torrentHash == "" {
		return errors.New("torrent was not added")
	}

	torrentHash, err = deluge.AddTorrentFile("ubuntu-14.04.6-desktop-amd64.iso.torrent", ubuntu14TorrentBase64, nil)
	if err != nil {
		return err
	}
	printServerResponse("AddTorrentFile", c)
	if torrentHash == "" {
		return errors.New("torrent was not added")
	}

	torrents, err := deluge.TorrentsStatus(delugeclient.StateUnspecified, nil)
	if err != nil {
		return err
	}
	printServerResponse("TorrentsStatus", c)

	found := false
	for id := range torrents {
		if id == testMagnetHash {
			found = true
			break
		}
	}
	if !found {
		return errors.New("cannot find torrent")
	}

	return nil
}

func printServerResponse(methodName string, c *delugeclient.Client) {
	if len(c.DebugServerResponses) != 1 {
		panic("BUG: expected exactly one response")
	}

	// store response for testing/development
	buf := c.DebugServerResponses[0]
	fmt.Printf("%s: received %d compressed bytes: %X\n", methodName, buf.Len(), buf.Bytes())

	c.DebugServerResponses = nil
}
