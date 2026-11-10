package test

import (
	"reflect"
	"testing"
	"time"
)

func AssertEqual[T any](t *testing.T, name string, want, got T) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func AssertNotEmpty(t *testing.T, name, got string) {
	t.Helper()
	if got == "" {
		t.Fatalf("%s is empty", name)
	}
}

func AssertNil(t *testing.T, name string, v any) {
	t.Helper()
	if !isNil(v) {
		t.Fatalf("%s expected nil, got %#v", name, v)
	}
}

func AssertTrue(t *testing.T, name string, v bool) {
	t.Helper()
	if !v {
		t.Fatalf("%s expected true, got false", name)
	}
}

func AssertNotZero[T any](t *testing.T, name string, got T) {
	t.Helper()
	if reflect.ValueOf(&got).Elem().IsZero() {
		t.Errorf("%s: expected non-zero value, got zero value (%#v)", name, got)
	}
}

func AssertTimeAfter(t *testing.T, name string, want, got time.Time) {
	t.Helper()
	if !got.After(want) {
		t.Fatalf("%s = %v, want after %v", name, got, want)
	}
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return rv.IsNil()
	}
	return false
}
