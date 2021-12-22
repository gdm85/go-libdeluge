// +build integration

package main

import (
	delugeclient "github.com/gdm85/go-libdeluge"
	"testing"
)

func TestLabelPluginGetLabels(t *testing.T) {
	var labelPlugin = &delugeclient.LabelPlugin{Client: c}

	_, err := labelPlugin.GetLabels()
	if err != nil {
		t.Fatal(err)
	}

	printServerResponse(t, "LabelPlugin.GetLabels")
}
