# kjkrol/uid

<p align="center">
  <img src=".github/docs/img/UID.png" alt="UID" width="300">
  <br>
  <a href="https://go.dev">
    <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  </a>
  <a href="https://pkg.go.dev/github.com/kjkrol/uid">
    <img src="https://img.shields.io/badge/GoDoc-Reference-007d9c?style=flat-square&logo=go" alt="GoDoc">
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/kjkrol/uid">
    <img src="https://goreportcard.com/badge/github.com/kjkrol/uid" alt="Go Report Card">
  </a>
  <a href="https://app.codecov.io/gh/kjkrol/uid">
    <img src="https://img.shields.io/codecov/c/github/kjkrol/uid?style=flat-square&logo=codecov" alt="Codecov Coverage">
  </a>
  <a href="https://github.com/kjkrol/uid/actions">
    <img src="https://github.com/kjkrol/uid/actions/workflows/go.yml/badge.svg" alt="Go Quality Check">
  </a>
</p>

A bit-packed 64-bit Generational Unique Identifier package for Go. 

Designed for memory pools and contiguous data structures. It provides index recycling with ABA problem prevention and an isolated 8-bit space for module-specific metadata.

## Overview

* **Zero-Allocation:** Identifier and bitmask manipulations are value-based.
* **Recycling:** Uses a generational counter to invalidate recycled pool indices.
* **Metadata Segments:** Provides a `MetaSegment` API to divide the 8-bit metadata space for safe bitwise operations across different modules.

## Installation

```bash
go get github.com/kjkrol/uid
```

## 64-bit Layout

The UID64 type is a uint64 partitioned into three segments:
```text
63        56 55                    32 31                             0
 +------------+------------------------+--------------------------------+
 |  Metadata  |       Generation       |             Index              |
 |  (8 bits)  |        (24 bits)       |            (32 bits)           |
 +------------+------------------------+--------------------------------+
```

- Index (32 bits): The main sequence number for array or pool offsets (Max: ~4.29 billion).
- Generation (24 bits): Incremented upon index release. Validates active references.
- Metadata (8 bits): User-defined space for boolean flags, enums, or module state.

## Usage

1. Pool Management (UID64Pool)
The pool handles allocation, generation increments, and index validation.

```go
package main

import (
	"fmt"
	"github.com/kjkrol/uid"
)

func main() {
	pool := uid.NewUID64Pool(1000, 100)

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
}
```

2. Custom Metadata (MetaSegment)

Divide the 8-bit Metadata space into strict, pre-shifted segments to prevent bit collisions between modules.

For example, defining a 1-bit segment at offset 7 (`virtualFlag`) and a 2-bit segment at offset 0 (`spatialFrag`):

```text
  7   6   5   4   3   2   1   0   (Metadata Bit Index)
+---+---+---+---+---+---+---+---+
| V |   |   |   |   |   | S | S |
+---+---+---+---+---+---+---+---+
  ^                       ^^^^^
  virtualFlag             spatialFrag
```

```go
package main

import (
	"fmt"
	"github.com/kjkrol/uid"
)

// Define segments globally per module
// Length: 1 bit, Offset: 7 (highest bit)
var virtualFlag = uid.NewMetaSegment(1, 7)

// Length: 2 bits, Offset: 0 (lowest bits)
var spatialFrag = uid.NewMetaSegment(2, 0)

func main() {
	id := uid.UID64(42) // Cast raw integer to UID64

	// Mutate segments
	id = id.WithMetaSegment(virtualFlag, 1)
	id = id.WithMetaSegment(spatialFrag, 3)

	// Read segments
	isVirtual := id.MetaSegment(virtualFlag) == 1
	fragValue := id.MetaSegment(spatialFrag)

	fmt.Printf("Virtual: %v, Frag: %d\n", isVirtual, fragValue)
}
```
