/*
 * go-libdeluge v0.3.1 - a native deluge RPC client library
 * Copyright (C) 2015~2019 gdm85 - https://github.com/gdm85/go-libdeluge/
This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gdm85/go-libdeluge"
)

var (
	host               string
	port               uint
	username, password string
	logLevel           string

	addURI       string
	listTorrents bool
	v2daemon     bool
	free         bool

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

	fs.BoolVar(&free, "f", false, "Display free space")
	fs.BoolVar(&free, "free", false, "Display free space")
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
		fmt.Fprintf(os.Stderr, "ERROR: invalid log level specified\n")
		os.Exit(2)
	}

	deluge := delugeclient.New(delugeclient.Settings{
		Hostname:              host,
		Port:                  port,
		Login:                 username,
		Password:              password,
		Logger:                logger,
		V2Daemon:              v2daemon,
		DebugSaveInteractions: debugIncoming})

	// perform connection to Deluge server
	err = deluge.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: connection failed: %v\n", err)
		os.Exit(3)
	}

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
		fmt.Printf("Free space: %d bytes\n", n)
	}

	if listTorrents {
		torrents, err := deluge.TorrentsStatus()

		// store response for testing/development
		count := len(deluge.DebugIncoming)
		if count != 0 {
			b := deluge.DebugIncoming[count-1]
			fmt.Println("last call received", len(b))
			err := ioutil.WriteFile("testlist.rnc", b, 0664)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: could not write last call test data: %v\n", err)
				os.Exit(5)
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not list all torrents: %v\n", err)
			os.Exit(6)
		}

		b, err := json.MarshalIndent(torrents, "", "\t")
		if err != nil {
			os.Exit(7)
		}
		fmt.Println(string(b))
	}
}
