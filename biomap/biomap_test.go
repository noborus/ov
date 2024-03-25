package biomap

import (
	"testing"
)

func TestMap(t *testing.T) {
	m := NewMap[int, string]()

	// Test Store and LoadForward
	m.Store(1, "one")
	value, ok := m.LoadForward(1)
	if !ok || value != "one" {
		t.Errorf("LoadForward failed. Expected value: %s, got: %s", "one", value)
	}

	// Test LoadBackward
	key, ok := m.LoadBackward("one")
	if !ok || key != 1 {
		t.Errorf("LoadBackward failed. Expected key: %d, got: %d", 1, key)
	}

	// Test DeleteForward
	m.DeleteForward(1)
	_, ok = m.LoadForward(1)
	if ok {
		t.Errorf("DeleteForward failed. Key still exists in Forward map")
	}

	// Test DeleteBackward
	m.DeleteBackward("one")
	_, ok = m.LoadBackward("one")
	if ok {
		t.Errorf("DeleteBackward failed. Value still exists in Backward map")
	}
}
