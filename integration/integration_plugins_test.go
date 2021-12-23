// +build integration

package main

import (
	delugeclient "github.com/gdm85/go-libdeluge"
	"testing"
	"time"
)

func waitForPluginEnabled(t *testing.T, name string) {
	tick := time.NewTicker(time.Millisecond * 500)
	defer tick.Stop()

	for attempt := 0; attempt < 10; attempt ++ {
		t.Logf("Attempt %d waiting for plugin %s to become enabled", attempt + 1, name)

		plugins, err := c.GetEnabledPlugins()
		if err != nil {
			t.Fatal(err)
		}

		for _, p := range plugins {
			if p == name {
				return
			}
		}

		// Sleep before trying again
		<-tick.C
	}

	t.Fatalf("Timeout waiting for plugin %s to become enabled", name)
}

func testWithPlugin(t *testing.T, name string) func() {
	err := deluge.EnablePlugin(name)
	if err != nil {
		t.Fatal(err)
	}

	waitForPluginEnabled(t, name)

	return func() {
		err := deluge.DisablePlugin(name)
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
