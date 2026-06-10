package uid

import (
	"testing"
)

// 1. Sequential Allocation
func TestUID64Pool_SequentialAllocation(t *testing.T) {
	pool := NewUID64Pool(10, 10)

	// Case: New entities receive incremental indices and generation 0
	for i := range 5 {
		e := pool.Next()
		index, gen := e.Unpack()

		if index != uint32(i) {
			t.Errorf("expected index %d, got %d", i, index)
		}
		if gen != 0 {
			t.Errorf("expected initial generation 0, got %d", gen)
		}
	}

	if pool.lastIndex != 5 {
		t.Errorf("expected lastIndex to be 5, got %d", pool.lastIndex)
	}
}

// 2. Index Recycling
func TestUID64Pool_Recycling(t *testing.T) {
	pool := NewUID64Pool(10, 10)

	// Case: Releasing and reusing an index
	e0 := pool.Next() // idx: 0, gen: 0
	pool.Release(e0)

	e0_new := pool.Next() // should be idx: 0, gen: 1
	index, gen := e0_new.Unpack()

	if index != 0 {
		t.Errorf("expected recycled index 0, got %d", index)
	}
	if gen != 1 {
		t.Errorf("expected incremented generation 1, got %d", gen)
	}

	// Case: LIFO behavior for free indices
	e1 := pool.Next() // idx: 1, gen: 0
	e2 := pool.Next() // idx: 2, gen: 0

	pool.Release(e1)
	pool.Release(e2)

	firstRecycled := pool.Next() // Should be e2 (last released)
	if firstRecycled.Index() != 2 {
		t.Errorf("expected LIFO recycling, got index %d, want 2", firstRecycled.Index())
	}
}

// 3. Validation Logic
func TestUID64Pool_Validation(t *testing.T) {
	pool := NewUID64Pool(10, 10)

	// Case: Active entity is valid
	e := pool.Next()
	if !pool.IsValid(e) {
		t.Error("expected active entity to be valid")
	}

	// Case: Released entity is invalid
	pool.Release(e)
	if pool.IsValid(e) {
		t.Error("expected released entity to be invalid (old generation)")
	}

	// Case: Out of bounds index
	futureUID := newUID(0, 999)
	if pool.IsValid(futureUID) {
		t.Error("expected out-of-bounds UID64 to be invalid")
	}

	// Case: Validating after recycling
	e_new := pool.Next() // Reuse index of 'e'
	if !pool.IsValid(e_new) {
		t.Error("expected newly recycled UID64 to be valid")
	}
	if pool.IsValid(e) {
		t.Error("old handle should remain invalid after recycling")
	}
}

// 4. Stress & Multiple Reuse
func TestUID64Pool_MultipleReuse(t *testing.T) {
	pool := NewUID64Pool(1, 1)
	idx := uint32(0)

	// Case: Cycling the same index multiple times
	for expectedGen := range uint32(10) {
		e := pool.Next()
		index, gen := e.Unpack()

		if index != idx || gen != expectedGen {
			t.Fatalf("iteration %d: expected idx %d gen %d, got idx %d gen %d",
				expectedGen, idx, expectedGen, index, gen)
		}
		pool.Release(e)
	}
}

// 5. Initial State
func TestUID64Pool_InitialState(t *testing.T) {
	capacity := 128
	pool := NewUID64Pool(capacity, capacity)

	// Case: Internal slices should respect initial capacity
	if cap(pool.generations) < capacity {
		t.Errorf("expected generations capacity at least %d, got %d", capacity, cap(pool.generations))
	}
	if len(pool.freeIndices) != 0 {
		t.Error("expected freeIndices to be empty initially")
	}
}

// 6. Dynamic Growth
func TestUID64Pool_Grow(t *testing.T) {
	// Case: Capacity zero grows to 8
	poolZero := NewUID64Pool(0, 0)
	poolZero.Next()
	if poolZero.capacity != 8 {
		t.Errorf("expected zero capacity to grow to 8, got %d", poolZero.capacity)
	}

	// Case: Capacity doubles when full
	pool := NewUID64Pool(2, 2)
	pool.Next() // idx 0
	pool.Next() // idx 1

	if pool.capacity != 2 {
		t.Errorf("expected capacity 2, got %d", pool.capacity)
	}

	pool.Next() // idx 2, triggers grow
	if pool.capacity != 4 {
		t.Errorf("expected capacity to double to 4, got %d", pool.capacity)
	}
	if cap(pool.generations) < 4 {
		t.Errorf("expected generations capacity at least 4, got %d", cap(pool.generations))
	}
}

// 7. Reset State
func TestUID64Pool_Reset(t *testing.T) {
	pool := NewUID64Pool(10, 10)

	// Case: Reset clears state
	e1 := pool.Next()
	pool.Release(e1)
	e2 := pool.Next() // generation 1

	pool.Reset()

	if pool.lastIndex != 0 {
		t.Errorf("expected lastIndex 0 after reset, got %d", pool.lastIndex)
	}
	if len(pool.freeIndices) != 0 {
		t.Errorf("expected freeIndices to be empty after reset")
	}

	// Case: Old entities are invalid
	if pool.IsValid(e2) {
		t.Error("expected previously active entity to be invalid after reset")
	}

	// Case: Allocates from beginning
	e3 := pool.Next()
	index, gen := e3.Unpack()
	if index != 0 || gen != 0 {
		t.Errorf("expected index 0 and gen 0 after reset, got idx %d gen %d", index, gen)
	}

	if pool.IsValid(e2) {
		t.Error("expected old entity to remain invalid even after reallocating its index")
	}
}
