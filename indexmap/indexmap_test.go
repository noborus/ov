package indexmap

import "testing"

func TestSetGetAndIndexOrder(t *testing.T) {
	m := NewIndexMap[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("a", 3) // update existing key, order must not change

	v, ok := m.Get("a")
	if !ok {
		t.Fatal("expected key a to exist")
	}
	if v != 3 {
		t.Fatalf("expected updated value 3, got %v", v)
	}

	k0, v0, ok := m.Index(0)
	if !ok {
		t.Fatal("expected index 0 to exist")
	}
	if k0 != "a" || v0 != 3 {
		t.Fatalf("expected index 0 to be (a,3), got (%v,%v)", k0, v0)
	}

	k1, v1, ok := m.Index(1)
	if !ok {
		t.Fatal("expected index 1 to exist")
	}
	if k1 != "b" || v1 != 2 {
		t.Fatalf("expected index 1 to be (b,2), got (%v,%v)", k1, v1)
	}

	if _, _, ok := m.Index(-1); ok {
		t.Fatal("expected index -1 to be out of range")
	}
	if _, _, ok := m.Index(2); ok {
		t.Fatal("expected index 2 to be out of range")
	}
}

func TestIteratorsOrderAndSnapshot(t *testing.T) {
	m := NewIndexMap[string, int]()
	m.Set("x", 10)
	m.Set("y", 20)

	all := m.All()
	keys := m.Keys()
	values := m.Values()

	// mutate after creating iterators to verify snapshot behavior
	m.Set("z", 30)
	m.Set("x", 99)

	var gotAll [][2]any
	for k, v := range all {
		gotAll = append(gotAll, [2]any{k, v})
	}
	if len(gotAll) != 2 {
		t.Fatalf("expected 2 items from All snapshot, got %d", len(gotAll))
	}
	if gotAll[0][0] != "x" || gotAll[0][1] != 10 {
		t.Fatalf("unexpected first pair from All: (%v,%v)", gotAll[0][0], gotAll[0][1])
	}
	if gotAll[1][0] != "y" || gotAll[1][1] != 20 {
		t.Fatalf("unexpected second pair from All: (%v,%v)", gotAll[1][0], gotAll[1][1])
	}

	var gotKeys []any
	for k := range keys {
		gotKeys = append(gotKeys, k)
	}
	if len(gotKeys) != 2 || gotKeys[0] != "x" || gotKeys[1] != "y" {
		t.Fatalf("unexpected Keys snapshot: %v", gotKeys)
	}

	var gotValues []any
	for v := range values {
		gotValues = append(gotValues, v)
	}
	if len(gotValues) != 2 || gotValues[0] != 10 || gotValues[1] != 20 {
		t.Fatalf("unexpected Values snapshot: %v", gotValues)
	}
}
