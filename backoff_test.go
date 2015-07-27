package cony

import (
	"math/rand"
	"testing"
)

func TestBackoffPolicy_Backoff(t *testing.T) {
	rand.Seed(0)
	bp := BackoffPolicy{[]int{0, 100, 1000, 10000}}

	if bp.Backoff(0) != 0 {
		t.Error("should return zero time duration")
	}

	tab := []struct {
		i  int
		ns int64
	}{
		{i: 1, ns: 124000000},
		{i: 2, ns: 1014000000},
		{i: 3, ns: 10353000000},
		{i: 4, ns: 7506000000},
	}

	for _, spec := range tab {
		ns := bp.Backoff(spec.i)
		if ns.Nanoseconds() != spec.ns {
			t.Errorf("try %d should produce %d nanoseconds, instead got %d",
				spec.i,
				spec.ns,
				ns,
			)
		}
	}
}
