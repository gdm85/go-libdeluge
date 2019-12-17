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
