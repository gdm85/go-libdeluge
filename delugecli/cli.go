// go-libdeluge v0.5.5 - a native deluge RPC client library
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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	delugeclient "github.com/gdm85/go-libdeluge"
)

var (
	host               string
	port               uint
	username, password string
	logLevel           string

	addURI               string
	listTorrents         bool
	listAvailablePlugins bool
	listEnabledPlugins   bool
	listAccounts         bool
	torrentHash          string
	setLabel             string
	addLabel             string
	removeLabel          string
	getLabels            bool
	listLabels           bool
	v2daemon             bool
	free                 bool
	testListenPort       bool
	sessionStatus        bool

	fs = flag.NewFlagSet("default", flag.ContinueOnError)
)

func init() {
	fs.StringVar(&host, "host", "localhost", "Deluge server host")
	fs.StringVar(&host, "h", "localhost", "Deluge server host (shorthand)")
	fs.UintVar(&port, "port", 58846, "Deluge server port")
	fs.UintVar(&port, "p", 58846, "Deluge server port (shorthand)")
	fs.StringVar(&username, "username", "localclient", "Deluge user name")
	fs.StringVar(&username, "u", "localclient", "Deluge user name (shorthand)")
	fs.StringVar(&password, "password", "", "Deluge password; use environment DELUGE_PASSWORD instead")
	fs.StringVar(&logLevel, "log-level", "", "Log level, one of 'DEBUG' or 'NONE'")
	fs.StringVar(&logLevel, "l", "", "Log level, one of 'DEBUG' or 'NONE' (shorthand)")

	fs.StringVar(&addURI, "a", "", "Add a torrent via magnet URI")
	fs.StringVar(&addURI, "add", "", "Add a torrent via magnet URI")

	fs.BoolVar(&v2daemon, "v2", false, "Use protocol compatible with a v2 daemon")

	fs.BoolVar(&listTorrents, "e", false, "List all torrents")
	fs.BoolVar(&listTorrents, "list", false, "List all torrents")
	fs.BoolVar(&listEnabledPlugins, "list-enabled-plugins", false, "List enabled plugins")
	fs.BoolVar(&listEnabledPlugins, "P", false, "List enabled plugins")
	fs.BoolVar(&listAvailablePlugins, "list-available-plugins", false, "List available plugins")
	fs.BoolVar(&listAvailablePlugins, "A", false, "List available plugins")

	fs.StringVar(&torrentHash, "torrent", "", "Operate on specified torrent hash")
	fs.StringVar(&torrentHash, "t", "", "Operate on specified torrent hash")
	fs.StringVar(&setLabel, "set-label", "", "Set label on torrent")
	fs.StringVar(&setLabel, "b", "", "Set label on torrent")
	fs.StringVar(&addLabel, "add-label", "", "Add label on torrent")
	fs.StringVar(&addLabel, "c", "", "Add label on torrent")
	fs.StringVar(&removeLabel, "remove-label", "", "Remove label on torrent")
	fs.StringVar(&removeLabel, "r", "", "Remove label on torrent")
	fs.BoolVar(&getLabels, "get-labels", false, "List all labels")
	fs.BoolVar(&listLabels, "list-labels", false, "List all torrents' labels")
	fs.BoolVar(&listLabels, "g", false, "List all torrents' labels")

	fs.BoolVar(&free, "f", false, "Display free space")
	fs.BoolVar(&free, "free", false, "Display free space")

	fs.BoolVar(&testListenPort, "o", false, "Test listen port")
	fs.BoolVar(&testListenPort, "test-listen-port", false, "Test listen port")

	fs.BoolVar(&listAccounts, "list-accounts", false, "List all known user accounts")
	fs.BoolVar(&sessionStatus, "s", false, "Show session status")
	fs.BoolVar(&sessionStatus, "session-status", false, "Show session status")
}

func main() {
	err := fs.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	// parse password via environment variable
	for _, decl := range os.Environ() {
		parts := strings.SplitN(decl, "=", 2)
		if len(parts) != 2 {
			continue
		}

		if parts[0] == "DELUGE_PASSWORD" {
			password = parts[1]
			break
		}
	}

	// validate log level
	var logger *log.Logger
	var debugIncoming bool
	switch logLevel {
	case "", "NONE":
		// nothing to do
	case "DEBUG":
		logger = log.New(os.Stderr, "DELUGE: ", log.Lshortfile)
		debugIncoming = true
	default:
		fmt.Fprintf(os.Stderr, "ERROR: invalid log level %q specified\n", logLevel)
		os.Exit(2)
	}

	settings := delugeclient.Settings{
		Hostname:             host,
		Port:                 port,
		Login:                username,
		Password:             password,
		Logger:               logger,
		DebugServerResponses: debugIncoming}

	deluge := delugeclient.NewV2(settings)

	// perform connection to Deluge server
	err = deluge.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
		os.Exit(3)
	}
	defer deluge.Close()

	// print daemon version
	ver, err := deluge.DaemonVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: daemon version retrieval: %v\n", err)
		os.Exit(4)
	}
	fmt.Printf("Deluge daemon version: %v\n", ver)

	// print available methods
	methods, err := deluge.MethodsList()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: methods list retrieval: %v\n", err)
		os.Exit(5)
	}
	fmt.Println("available methods:", methods)

	// add each of the torrents
	if addURI != `` {
		torrentHash, err := deluge.AddTorrentMagnet(addURI, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not add magnet URI '%s': %v\n", addURI, err)
			os.Exit(5)
		}

		if torrentHash == "" {
			fmt.Println("torrent was not added")
		} else {
			fmt.Println("added torrent with hash:", torrentHash)
		}

	}

	if free {
		n, err := deluge.GetFreeSpace("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not read free space: %v\n", err)
			os.Exit(6)
		}
		fmt.Printf("free space: %d bytes\n", n)
	}

	if listAvailablePlugins {
		plugins, err := deluge.GetAvailablePlugins()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: available plugins list retrieval: %v\n", err)
			os.Exit(5)
		}
		fmt.Println("available plugins:", plugins)
	}

	if listEnabledPlugins {
		plugins, err := deluge.GetEnabledPlugins()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: enabled plugins list retrieval: %v\n", err)
			os.Exit(5)
		}
		fmt.Println("enabled plugins:", plugins)
	}

	if setLabel != "" {
		if torrentHash == "" {
			fmt.Fprintf(os.Stderr, "ERROR: no torrent hash specified\n")
			os.Exit(5)
		}
		p, err := deluge.LabelPlugin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: enabled plugins list retrieval: %v\n", err)
			os.Exit(5)
		}
		err = p.SetTorrentLabel(torrentHash, setLabel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: setting label %q on torrent %q: %v\n", setLabel, torrentHash, err)
			os.Exit(5)
		}
	}

	if addLabel != "" {
		if torrentHash != "" {
			fmt.Fprintf(os.Stderr, "ERROR: no torrent hash should be specified\n")
			os.Exit(5)
		}
		p, err := deluge.LabelPlugin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: label plugin: %v\n", err)
			os.Exit(5)
		}
		err = p.AddLabel(addLabel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: adding label %q: %v\n", addLabel, err)
			os.Exit(5)
		}
	}

	if removeLabel != "" {
		if torrentHash != "" {
			fmt.Fprintf(os.Stderr, "ERROR: no torrent hash should be specified\n")
			os.Exit(5)
		}
		p, err := deluge.LabelPlugin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: label plugin: %v\n", err)
			os.Exit(5)
		}
		err = p.RemoveLabel(removeLabel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: removing label %q: %v\n", addLabel, err)
			os.Exit(5)
		}
	}

	if getLabels {
		p, err := deluge.LabelPlugin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: label plugin: %v\n", err)
			os.Exit(5)
		}

		labels, err := p.GetLabels()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: label plugin: %v\n", err)
			os.Exit(5)
		}

		je := json.NewEncoder(os.Stdout)
		je.SetIndent("", "\t")
		if err := je.Encode(labels); err != nil {
			os.Exit(7)
		}
	}

	if listLabels {
		p, err := deluge.LabelPlugin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: enabled plugins list retrieval: %v\n", err)
			os.Exit(5)
		}
		labelsByTorrent, err := p.GetTorrentsLabels(delugeclient.StateUnspecified, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: enabled plugins list retrieval: %v\n", err)
			os.Exit(5)
		}
		je := json.NewEncoder(os.Stdout)
		je.SetIndent("", "\t")
		if err := je.Encode(labelsByTorrent); err != nil {
			os.Exit(7)
		}
	}

	if listTorrents {
		torrents, err := deluge.TorrentsStatus(delugeclient.StateUnspecified, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not list all torrents: %v\n", err)
			os.Exit(6)
		}

		je := json.NewEncoder(os.Stdout)
		je.SetIndent("", "\t")
		if err := je.Encode(torrents); err != nil {
			os.Exit(7)
		}
	}

	if listAccounts {
		accounts, err := deluge.KnownAccounts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not list all accounts: %v\n", err)
			os.Exit(6)
		}

		je := json.NewEncoder(os.Stdout)
		je.SetIndent("", "\t")
		if err := je.Encode(accounts); err != nil {
			os.Exit(7)
		}
	}

	if testListenPort {
		success, err := deluge.TestListenPort()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not test listen port: %v\n", err)
			os.Exit(6)
		}
		fmt.Printf("test listen port: %v\n", success)
	}

	if sessionStatus {
		status, err := deluge.GetSessionStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not get session status: %v\n", err)
			os.Exit(6)
		}
		fmt.Printf("session status: %+v\n", status)
	}
}
