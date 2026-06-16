// Package uid implements UID64, a bit-packed 64-bit generational identifier for
// memory pools and contiguous, index-addressed data structures such as entity
// systems, slab allocators and slot maps.
//
// # Layout
//
// A UID64 packs three fields into a single uint64:
//
//		63        56 55                    32 31                             0
//		+------------+------------------------+--------------------------------+
//		|  Metadata  |       Generation       |             Index              |
//		|  (8 bits)  |        (24 bits)       |            (32 bits)           |
//		+------------+------------------------+--------------------------------+
//
//	  - Index (32 bits): slot offset into a backing array (up to ~4.29 billion).
//	  - Generation (24 bits): bumped each time a slot is released, so a stale
//	    reference to a recycled slot fails validation — this prevents the ABA
//	    problem.
//	  - Metadata (8 bits): user-defined space, partitioned with [MetaSegment].
//
// All extraction and mutation is value-based and allocation-free.
//
// # Pools
//
// [UID64Pool] hands out identifiers, recycles released indices and validates
// references. Initialise it in place with [UID64Pool.Init]:
//
//	var pool uid.UID64Pool
//	pool.Init(1000, 100)
//
//	id := pool.Next()        // allocate one
//	pool.Release(id)         // recycle the index; id no longer validates
//	next := pool.Next()      // reuses the index under a bumped generation
//
// [UID64Pool.NextN] reserves many identifiers at once into a caller-owned
// buffer, amortising the per-id branch, capacity check and growth — preferable
// for bulk creation.
//
// # Metadata segments
//
// [MetaSegment] divides the 8-bit metadata space into fixed, pre-shifted
// bit-ranges so independent modules can store flags without colliding. See
// [NewMetaSegment], [UID64.WithMetaSegment] and [UID64.MetaSegment].
package uid
