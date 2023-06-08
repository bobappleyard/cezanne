package assert

import (
	"reflect"
	"testing"
)

func Equal[T any](t testing.TB, got, expected T) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, expected %#v", got, expected)
	}
}

func Nil(t testing.TB, got any) {
	t.Helper()
	if got == nil {
		return
	}
	if reflect.ValueOf(got).IsNil() {
		return
	}
	t.Errorf("got %v, expecting nil", got)
}

func True(t testing.TB, got bool) {
	t.Helper()
	Equal(t, got, true)
}

func False(t testing.TB, got bool) {
	t.Helper()
	Equal(t, got, false)
}
