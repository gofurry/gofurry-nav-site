package service

import "testing"

func TestHSetMapSkipsEmptyMap(t *testing.T) {
	if err := HSetMap("test:empty", map[string]string{}); err != nil {
		t.Fatalf("HSetMap() should skip empty maps without Redis command, got %v", err)
	}

	if err := HSetMap("test:nil", nil); err != nil {
		t.Fatalf("HSetMap() should skip nil maps without Redis command, got %v", err)
	}
}
