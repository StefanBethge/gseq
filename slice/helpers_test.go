package slice

import (
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("mismatch\nwant: %v\n got: %v", want, got)
	}
}
