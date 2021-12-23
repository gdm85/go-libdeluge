// +build integration

package main

import (
	delugeclient "github.com/gdm85/go-libdeluge"
	"testing"
)

func testWithPlugin(t *testing.T, name string) func() {
	_, err := deluge.EnablePlugin(name)
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		_, err := deluge.DisablePlugin(name)
		if err != nil {
			t.Fatal(err)
		}

		// cleanup DebugServerResponses after plugin disable
		c.DebugServerResponses = nil
	}
}

func TestLabelPluginGetLabels(t *testing.T) {
	defer testWithPlugin(t, "Label")

	var labelPlugin = &delugeclient.LabelPlugin{Client: c}

	_, err := labelPlugin.GetLabels()
	if err != nil {
		t.Fatal(err)
	}

	printServerResponse(t, "LabelPlugin.GetLabels")
}
