// +build integration,integration_v2,!integration_v1

package main

import (
	"testing"

	delugeclient "github.com/gdm85/go-libdeluge"
)

var (
	v2daemon = true
)

func TestKnownAccounts(t *testing.T) {
	if !v2daemon {
		t.Skip()
		return
	}

	deluge := deluge.(delugeclient.V2)
	_, err := deluge.KnownAccounts()
	if err != nil {
		t.Fatal(err)
	}
	printServerResponse("KnownAccounts", c)
}
