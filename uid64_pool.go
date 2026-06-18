package uid

// UID64Pool manages the allocation, recycling, and validation of UID64 identifiers.
type UID64Pool struct {
	nextIndex   uint32
	freeIndices []uint32
	generations []uint32
	capacity    uint32
}

// Init prepares the pool with pre-allocated capacities: indexCap sizes the index
// generation table (the index space the pool can address before growing) and
// recycleCap sizes the released-index recycling list. It resets any prior state,
// so it may also be used to re-initialise a reused pool.
func (p *UID64Pool) Init(indexCap, recycleCap int) {
	p.nextIndex = 0
	p.freeIndices = make([]uint32, 0, recycleCap)
	p.generations = make([]uint32, indexCap)
	p.capacity = uint32(indexCap)
}

// Reset clears the pool state, invalidating all previously issued UID64s.
func (p *UID64Pool) Reset() {
	p.nextIndex = 0
	p.freeIndices = p.freeIndices[:0]
	clear(p.generations)
}

// Next returns a new, valid UID64. It reuses an old index if available,
// otherwise it allocates a new one.
func (p *UID64Pool) Next() UID64 {
	if fLen := len(p.freeIndices); fLen > 0 {
		index := p.freeIndices[fLen-1]
		p.freeIndices = p.freeIndices[:fLen-1]
		gen := p.generations[index]
		return newUID(gen, index)
	}

	p.ensure(p.nextIndex + 1)

	index := p.nextIndex
	p.nextIndex++

	return newUID(p.generations[index], index)
}

// NextN reserves len(dst) fresh UID64s and writes them into dst. The caller
// owns the buffer, so it is zero-alloc and reusable across calls. Recycled
// indices are drained first (LIFO, with their current generation), then the
// remainder is allocated as one contiguous run from the high-water mark,
// growing the generation table at most once — amortising the per-id branch,
// capacity check and growth that Next() pays on every call.
func (p *UID64Pool) NextN(dst []UID64) {
	w := 0

	for w < len(dst) && len(p.freeIndices) > 0 {
		fLen := len(p.freeIndices)
		index := p.freeIndices[fLen-1]
		p.freeIndices = p.freeIndices[:fLen-1]
		dst[w] = newUID(p.generations[index], index)
		w++
	}

	if remaining := len(dst) - w; remaining > 0 {
		p.ensure(p.nextIndex + uint32(remaining))
		for ; w < len(dst); w++ {
			index := p.nextIndex
			p.nextIndex++
			dst[w] = newUID(p.generations[index], index)
		}
	}
}

// PeekNextIndex returns the index that will be assigned to the next sequentially
// allocated UID64. It does not reserve or advance any state. Use after NextN to
// obtain the sequential high-water mark for capacity checks.
func (p *UID64Pool) PeekNextIndex() uint32 {
	return p.nextIndex
}

// ensure guarantees the generation table can address indices up to need-1,
// growing by doubling (and at least to need) when short.
func (p *UID64Pool) ensure(need uint32) {
	if p.capacity >= need {
		return
	}

	newCap := p.capacity * 2
	if newCap == 0 {
		newCap = 8
	}
	for newCap < need {
		newCap *= 2
	}

	newGenerations := make([]uint32, newCap)
	copy(newGenerations, p.generations)
	p.generations = newGenerations
	p.capacity = newCap
}

// Release invalidates the given UID64 and recycles its index for future use.
// It returns the underlying index.
func (p *UID64Pool) Release(u UID64) uint32 {
	index := u.Index()

	p.generations[index] = (p.generations[index] + 1) & GenerationMask
	p.freeIndices = append(p.freeIndices, index)

	return index
}

// IsValid verifies if the given UID64 is currently active in the pool.
func (p *UID64Pool) IsValid(u UID64) bool {
	index, gen := u.Unpack()
	return index < p.nextIndex && p.generations[index] == gen
}
