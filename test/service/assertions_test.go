package service_test

import (
	"reflect"
	"testing"
)

func assertEqual[T any](t *testing.T, name string, want, got T) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}
