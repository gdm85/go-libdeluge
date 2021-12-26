// +build integration

package main

import (
	delugeclient "github.com/gdm85/go-libdeluge"
	"testing"
)


func enablePlugin(t *testing.T, name string) {
	err := deluge.EnablePlugin(name)
	if err != nil {
		t.Fatal(err)
	}
}

func disablePlugin(t *testing.T, name string) {
	err := deluge.DisablePlugin(name)
	if err != nil {
		t.Fatal(err)
	}

	// cleanup DebugServerResponses after plugin disable
	c.DebugServerResponses = nil
}

func TestLabelPluginGetLabels(t *testing.T) {
	enablePlugin(t, "Label")
	defer disablePlugin(t, "Label")

	var labelPlugin = &delugeclient.LabelPlugin{Client: c}

	_, err := labelPlugin.GetLabels()
	if err != nil {
		t.Fatal(err)
	}

	printServerResponse(t, "LabelPlugin.GetLabels")
}
