package uid

// UID64Pool manages the allocation, recycling, and validation of UID64 identifiers.
type UID64Pool struct {
	lastIndex   uint32
	freeIndices []uint32
	generations []uint32
	capacity    uint32
}

// NewUID64Pool creates a new generational pool for UID64 with pre-allocated capacities.
func NewUID64Pool(initialCap, freeIndicesCap int) UID64Pool {
	return UID64Pool{
		generations: make([]uint32, initialCap),
		freeIndices: make([]uint32, 0, freeIndicesCap),
		capacity:    uint32(initialCap),
	}
}

// Reset clears the pool state, invalidating all previously issued UID64s.
func (p *UID64Pool) Reset() {
	p.lastIndex = 0
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
		return New(gen, index)
	}

	if p.lastIndex >= p.capacity {
		p.grow()
	}

	index := p.lastIndex
	p.lastIndex++

	return New(p.generations[index], index)
}

func (p *UID64Pool) grow() {
	newCap := p.capacity * 2
	if newCap == 0 {
		newCap = 8
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
	return index < p.lastIndex && p.generations[index] == gen
}
