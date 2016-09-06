package uuid

import (
	"testing"
)

func Test_NewUUID(t *testing.T) {
	if u, err := NewUUID(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(u)
	}
}
