package dao

import (
	"testing"
)

func Test_InitDAO(t *testing.T) {
	openDB(t)
	if err := InitDAO(); err != nil {
		t.Fatal(err)
	}
}
