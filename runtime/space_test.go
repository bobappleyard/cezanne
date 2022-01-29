package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpacePopulate(t *testing.T) {

	var s Space

	o1 := s.Class(1, []MethodID{3, 1, 4})
	o2 := s.Class(2, []MethodID{7, 4, 6})
	o3 := s.Class(3, []MethodID{12})

	assertFoundOffset(t, &s, 1, o1, 3, 0)
	assertFoundOffset(t, &s, 1, o1, 1, 1)
	assertFoundOffset(t, &s, 1, o1, 4, 2)

	assertFoundOffset(t, &s, 2, o2, 7, 0)
	assertFoundOffset(t, &s, 2, o2, 4, 1)
	assertFoundOffset(t, &s, 2, o2, 6, 2)

	assertFoundOffset(t, &s, 3, o3, 12, 0)

	assertMissingMethod(t, &s, 1, o1, 0)
	assertMissingMethod(t, &s, 1, o1, 2)
	assertMissingMethod(t, &s, 1, o1, 7)

	assertMissingMethod(t, &s, 2, o2, 0)
	assertMissingMethod(t, &s, 2, o2, 2)
	assertMissingMethod(t, &s, 2, o2, 8)

}

func assertFoundOffset(t *testing.T, s *Space, class ClassID, off int, method MethodID, index int) {
	t.Helper()
	idx, err := s.LookupMethod(method, class, off)

	assert.Nil(t, err)
	assert.Equal(t, index, idx)
}

func assertMissingMethod(t *testing.T, s *Space, class ClassID, off int, method MethodID) {
	t.Helper()
	_, err := s.LookupMethod(method, class, off)

	assert.Equal(t, ErrUnknownMember, err)
}
