// go-libdeluge v0.3.1 - a native deluge RPC client library
// Copyright (C) 2015~2019 gdm85 - https://github.com/gdm85/go-libdeluge/
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

var testOpts = Options{
	MaxConnections:     100,
	AutoManaged:        true,
	PreAllocateStorage: true,
	V2: V2Options{
		Shared: true,
	},
}

func TestNilOptionsEncode(t *testing.T) {
	t.Parallel()
	var o *Options
	d := o.toDictionary(false)
	if d.Length() != 0 {
		t.Error("expected an empty dictionary")
	}
}

func TestOptionsEncodeV1(t *testing.T) {
	t.Parallel()

	d := testOpts.toDictionary(false)

	m, err := d.Zip()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := m["compact_allocation"]; !ok {
		t.Errorf("expected key %q not found", "compact_allocation")
	}

	if _, ok := m["shared"]; ok {
		t.Errorf("unexpected key %q found", "shared")

	}
}

func TestOptionsEncodeV2(t *testing.T) {
	t.Parallel()

	d := testOpts.toDictionary(true)

	m, err := d.Zip()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := m["compact_allocation"]; ok {
		t.Errorf("unexpected key %q found", "compact_allocation")
	}

	if _, ok := m["pre_allocate_storage"]; !ok {
		t.Errorf("expected key %q not found", "pre_allocate_storage")

	}

	if _, ok := m["shared"]; !ok {
		t.Errorf("expected key %q not found", "shared")

	}
}
