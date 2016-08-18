/*
 * go-libdeluge v0.1.0 - a native deluge RPC client library
 * Copyright (C) 2015~2016 gdm85 - https://github.com/gdm85/go-libdeluge/
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
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdm85/go-libdeluge"
)

var (
	host               string
	port               uint
	username, password string
	logLevel           string

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
}

func main() {
	err := fs.Parse(os.Args[1:])
	if err != nil {
		//		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// make sure there is at least one magnet URI specified after CLI flags
	if fs.NArg() == 0 {
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "ERROR: at least one magnet URI should be specified\n")
		os.Exit(1)
	}

	// parse options via environment variables
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
	switch logLevel {
	case "", "NONE":
		// nothing to do
	case "DEBUG":
		logger = log.New(os.Stderr, "DELUGE: ", log.Lshortfile)
	default:
		fmt.Fprintf(os.Stderr, "ERROR: invalid log level specified\n")
		os.Exit(2)
	}

	deluge := delugeclient.New(delugeclient.Settings{host, port, username, password, logger, time.Duration(0)})

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
	for i, magnetUri := range fs.Args() {
		torrentHash, err := deluge.AddTorrentMagnet(magnetUri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not add magnet URI '%s': %v\n", magnetUri, err)
			os.Exit(5)
		}

		if torrentHash == "" {
			fmt.Printf("torrent #%d was not added\n", i)
		} else {
			fmt.Printf("added torrent #%d with hash: %s\n", i, torrentHash)
		}

	}
}
