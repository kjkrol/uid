package uid

import (
	"fmt"
	"testing"
)

// TestNextN_MatchesNextLoop asserts NextN yields exactly what a sequence of
// Next() calls would, for a cold pool (pure fresh allocation).
func TestNextN_MatchesNextLoop(t *testing.T) {
	const n = 1000

	var a UID64Pool
	a.Init(0, 0)
	want := make([]UID64, n)
	for i := range want {
		want[i] = a.Next()
	}

	var b UID64Pool
	b.Init(0, 0)
	got := make([]UID64, n)
	b.NextN(got)

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: NextN=%#x, Next-loop=%#x", i, got[i], want[i])
		}
	}
}

// TestNextN_DrainsRecycledFirst asserts NextN reuses released indices (with
// bumped generation) in the same order Next() would, before allocating fresh.
func TestNextN_DrainsRecycledFirst(t *testing.T) {
	build := func(useBatch bool) []UID64 {
		var p UID64Pool
		p.Init(0, 0)
		seed := make([]UID64, 10)
		p.NextN(seed)
		p.Release(seed[3])
		p.Release(seed[7])

		out := make([]UID64, 5) // 2 recycled + 3 fresh
		if useBatch {
			p.NextN(out)
		} else {
			for i := range out {
				out[i] = p.Next()
			}
		}
		return out
	}

	loop, batch := build(false), build(true)
	for i := range loop {
		if loop[i] != batch[i] {
			t.Fatalf("index %d: batch=%#x, loop=%#x", i, batch[i], loop[i])
		}
	}
}

// BenchmarkPoolNext compares one-by-one Next() against batched NextN().
//   - warm: capacity pre-sized, isolating per-id branch/call overhead.
//   - cold: empty pool, including the growth that NextN performs once vs the
//     log2(n) doublings a Next() loop triggers.
func BenchmarkPoolNext(b *testing.B) {
	for _, n := range []int{1, 100, 10000} {
		dst := make([]UID64, n)

		b.Run(fmt.Sprintf("warm/one-by-one/%d", n), func(b *testing.B) {
			var p UID64Pool
			p.Init(n, 0)
			for b.Loop() {
				p.lastIndex = 0
				for i := range dst {
					dst[i] = p.Next()
				}
			}
		})
		b.Run(fmt.Sprintf("warm/batch/%d", n), func(b *testing.B) {
			var p UID64Pool
			p.Init(n, 0)
			for b.Loop() {
				p.lastIndex = 0
				p.NextN(dst)
			}
		})

		b.Run(fmt.Sprintf("cold/one-by-one/%d", n), func(b *testing.B) {
			for b.Loop() {
				var p UID64Pool
				p.Init(0, 0)
				for i := range dst {
					dst[i] = p.Next()
				}
			}
		})
		b.Run(fmt.Sprintf("cold/batch/%d", n), func(b *testing.B) {
			for b.Loop() {
				var p UID64Pool
				p.Init(0, 0)
				p.NextN(dst)
			}
		})
	}
}
