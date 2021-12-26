package delugeclient

import "testing"

func TestLabelPlugin_GetLabels(t *testing.T) {
	t.Parallel()

	c := newMockClient(1, "789C3BCCC874A839B338BF29BF18001B7504BB")

	plugin := &LabelPlugin{
		Client: c.(*Client),
	}

	labels, err := plugin.GetLabels()
	if err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"iso",
		"os",
	}

	if len(labels) != len(expect) {
		t.Fatalf("mismatch in expected label count: expected %d, got %d", len(expect), len(labels))
	}

	for i := 0; i < len(labels); i++ {
		expectLabel := expect[i]
		givenLabel := labels[i]

		if expectLabel != givenLabel {
			t.Fatalf("wrong label: expected %s; got %s", expectLabel, givenLabel)
		}
	}
}
