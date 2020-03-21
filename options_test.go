// go-libdeluge v0.5.0 - a native deluge RPC client library
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

package delugeclient

import (
	"testing"
)

var testOpts Options

func init() {
	maxConns := 100
	tVal := true
	testOpts.MaxConnections = &maxConns
	testOpts.AutoManaged = &tVal
	testOpts.PreAllocateStorage = &tVal
	testOpts.V2.Shared = &tVal
}

func TestNilOptionsEncode(t *testing.T) {
	t.Parallel()
	var o *Options
	d := o.toDictionary(false)
	if d.Length() != 0 {
		t.Error("expected an empty dictionary")
	}
}

func TestDefaultEncode(t *testing.T) {
	testOptsWithDefaults := testOpts
	fVal := false
	testOptsWithDefaults.StopAtRatio = &fVal

	d := testOptsWithDefaults.toDictionary(false)

	m, err := d.Zip()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encoded options: %#v\n", m)

	if _, ok := m["stop_at_ratio"]; !ok {
		t.Errorf("expected key %q not found", "stop_at_ratio")
	}
}

func TestOptionsEncodeV1(t *testing.T) {
	t.Parallel()

	d := testOpts.toDictionary(false)

	m, err := d.Zip()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encoded options: %#v\n", m)

	if _, ok := m["compact_allocation"]; !ok {
		t.Errorf("expected key %q not found", "compact_allocation")
	}

	if _, ok := m["shared"]; ok {
		t.Errorf("unexpected key %q found", "shared")

	}

	// a field never specified should not be encoded
	if _, ok := m["max_upload_slots"]; ok {
		t.Errorf("unexpected key %q found", "max_upload_slots")

	}
}

func TestOptionsEncodeV2(t *testing.T) {
	t.Parallel()

	d := testOpts.toDictionary(true)

	m, err := d.Zip()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encoded options: %#v\n", m)

	if _, ok := m["compact_allocation"]; ok {
		t.Errorf("unexpected key %q found", "compact_allocation")
	}

	if _, ok := m["pre_allocate_storage"]; !ok {
		t.Errorf("expected key %q not found", "pre_allocate_storage")

	}

	if _, ok := m["shared"]; !ok {
		t.Errorf("expected key %q not found", "shared")

	}

	// a field never specified should not be encoded
	if _, ok := m["max_upload_slots"]; ok {
		t.Errorf("unexpected key %q found", "max_upload_slots")

	}
}
