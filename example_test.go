package uid_test

import (
	"fmt"

	"github.com/kjkrol/uid"
)

// ExampleUID64Pool shows the allocate / release / recycle lifecycle: releasing
// an id bumps its index's generation, so the old reference stops validating and
// the next allocation reuses the index under a fresh generation.
func ExampleUID64Pool() {
	var pool uid.UID64Pool
	pool.Init(1000, 100)

	// Allocate
	id := pool.Next()

	// Unpack returns both index and generation
	index, gen := id.Unpack()
	fmt.Printf("Allocated: Index %d, Gen %d\n", index, gen)

	// Release invalidates the current generation of this index
	pool.Release(id)

	if !pool.IsValid(id) {
		fmt.Println("Old reference is invalid.")
	}

	// Next allocation reuses the index but increments the generation
	recycled := pool.Next()
	fmt.Printf("Recycled: Index %d, Gen %d\n", recycled.Index(), recycled.Generation())

	// Output:
	// Allocated: Index 0, Gen 0
	// Old reference is invalid.
	// Recycled: Index 0, Gen 1
}

// ExampleUID64Pool_batch shows NextN reserving several ids at once into a
// caller-owned buffer (zero allocation).
func ExampleUID64Pool_batch() {
	var pool uid.UID64Pool
	pool.Init(1000, 100)

	ids := make([]uid.UID64, 4)
	pool.NextN(ids)

	for _, id := range ids {
		fmt.Printf("Index %d, Gen %d\n", id.Index(), id.Generation())
	}

	// Output:
	// Index 0, Gen 0
	// Index 1, Gen 0
	// Index 2, Gen 0
	// Index 3, Gen 0
}
