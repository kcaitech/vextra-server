package set

import (
	"testing"
)

func Test0(t *testing.T) {
	set := NewSet(1, 2, 3, 4, 5)
	set.Remove(3)
	if set.Contains(3) {
		t.Error("set.Contains(3) = true, want false")
	}
	if set.Size() != 4 {
		t.Errorf("set.Size() = %d, want 4", set.Size())
	}
	set.Clear()
	if set.Size() != 0 {
		t.Errorf("set.Size() = %d, want 0", set.Size())
	}
}

func Test1(t *testing.T) {
	var (
		a = byte(0x00)
		b = byte(0x02)
	)
	set := NewSet(a, b)
	if !set.Contains(a) {
		t.Error("set.Contains(a) = false, want true")
	}
}
