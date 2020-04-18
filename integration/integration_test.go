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
	"fmt"
	"log"
	"os"
	"testing"

	delugeclient "github.com/gdm85/go-libdeluge"
)

const (
	testMagnetHash = "c1939ca413b9afcc34ea0cf3c128574e93ff6cb0"
	testMagnetURI  = `magnet:?xt=urn:btih:c1939ca413b9afcc34ea0cf3c128574e93ff6cb0&tr=http%3A%2F%2Ftorrent.ubuntu.com%3A6969%2Fannounce`
)

var (
	deluge   delugeclient.DelugeClient
	c        *delugeclient.Client
	settings = delugeclient.Settings{
		Hostname: "127.0.0.1",
		Port:     58846,
		Login:    "localclient",
		Password: "deluge",
	}
)

func TestMain(m *testing.M) {
	err := prepareClient(settings)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	deluge.Close()

	os.Exit(exitCode)
}

func prepareClient(settings delugeclient.Settings) error {
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
		return fmt.Errorf("connection failed: %w", err)
	}
	printServerResponse("DaemonLogin", c)

	return nil
}

func TestDaemonVersion(t *testing.T) {
	_, err := deluge.DaemonVersion()
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("DaemonVersion", c)
}

func TestMethodsList(t *testing.T) {
	methods, err := deluge.MethodsList()
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("MethodsList", c)
	if len(methods) == 0 {
		t.Error("no methods returned")
	}
}

func TestFreeSpace(t *testing.T) {
	_, err := deluge.GetFreeSpace("")
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("GetFreeSpace", c)
}

func TestGetAvailablePlugins(t *testing.T) {
	_, err := deluge.GetAvailablePlugins()
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("GetAvailablePlugins", c)
}

func TestGetEnabledPlugins(t *testing.T) {
	_, err := deluge.GetEnabledPlugins()
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("GetEnabledPlugins", c)
}

func TestAddTorrentMagnet(t *testing.T) {
	torrentHash, err := deluge.AddTorrentMagnet(testMagnetURI, nil)
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("AddTorrentMagnet", c)
	if torrentHash == "" {
		t.Error("torrent was not added")
	}
}

func TestAddTorrentFile(t *testing.T) {
	torrentHash, err := deluge.AddTorrentFile("ubuntu-14.04.6-desktop-amd64.iso.torrent", ubuntu14TorrentBase64, nil)
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("AddTorrentFile", c)
	if torrentHash == "" {
		t.Error("torrent was not added")
	}
}

func TestTorrentsStatus(t *testing.T) {
	torrents, err := deluge.TorrentsStatus(delugeclient.StateUnspecified, nil)
	if err != nil {
		t.Fatal(err)
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
		t.Error("cannot find torrent")
	}
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
