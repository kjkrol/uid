package uid

const (
	// IndexMask isolates the lower 32 bits used for the entity index.
	IndexMask = 0xFFFFFFFF

	// GenerationShift is the bit offset for the generation counter.
	GenerationShift = 32

	// GenerationMask isolates the 24 bits allocated for the generation counter.
	GenerationMask = 0xFFFFFF

	// MetadataShift is the bit offset for the custom metadata byte.
	MetadataShift = 56

	// MetadataMask isolates the upper 8 bits allocated for user-defined metadata.
	MetadataMask = 0xFF
)

// MetaSegment defines a strictly bounded bit-range within the 8-bit Metadata space.
// It stores a pre-shifted mask for zero-cost bit clears during updates.
type MetaSegment struct {
	mask  uint8 // Pre-shifted mask for exact positioning
	shift uint8 // Necessary for normalizing the extracted value
}

// NewMetaSegment creates a definition for a metadata field.
//   - length: the number of bits allocated to this segment (1 to 8).
//   - offset: the starting bit position from the right (0 to 7).
//
// Example of a segment with length=3 and offset=2:
//
//	  7   6   5   4   3   2   1   0   (Metadata Bit Index)
//	+---+---+---+---+---+---+---+---+
//	|   |   |   | ■ | ■ | ■ |   |   |
//	+---+---+---+---+---+---+---+---+
//	              ^^^^^^^^^
//	            Segment Space
func NewMetaSegment(length, offset uint8) MetaSegment {
	unshiftedMask := uint8((1 << length) - 1)
	return MetaSegment{
		mask:  unshiftedMask << offset,
		shift: offset,
	}
}

// UID64 is a bit-packed 64-bit Generational Unique Identifier.
//
// Bit layout configuration:
//
//	 63        56 55                    32 31                             0
//	+------------+------------------------+--------------------------------+
//	|  Metadata  |       Generation       |             Index              |
//	|  (8 bits)  |        (24 bits)       |            (32 bits)           |
//	+------------+------------------------+--------------------------------+
//
// Component breakdown:
//   - Bits 00-31: Index (32 bits) - Main sequence for array/pool offsets.
//   - Bits 32-55: Generation (24 bits) - Recycling counter preventing ABA problems.
//   - Bits 56-63: Metadata (8 bits) - User-defined space for module flags.
type UID64 uint64

// newUID creates a new UID64 from the provided generation and index.
func newUID(gen, index uint32) UID64 {
	gen &= GenerationMask
	return UID64(uint64(gen)<<GenerationShift | uint64(index))
}

// Index extracts the 32-bit index component from the UID64.
func (u UID64) Index() uint32 {
	return uint32(u & IndexMask)
}

// Generation extracts the 24-bit generation counter from the UID64.
func (u UID64) Generation() uint32 {
	return uint32((u >> GenerationShift) & GenerationMask)
}

// Unpack extracts both the index and the generation from the UID64.
func (u UID64) Unpack() (uint32, uint32) {
	return uint32(u & IndexMask), uint32((u >> GenerationShift) & GenerationMask)
}

// MetaSegment reads the value of a specific metadata segment directly from the UID64.
func (u UID64) MetaSegment(s MetaSegment) uint8 {
	return (u.metadata() & s.mask) >> s.shift
}

// WithMetaSegment returns a copy of the UID64 with only the specified segment updated.
// Excess bits in the provided value are silently truncated to prevent corruption.
func (u UID64) WithMetaSegment(s MetaSegment, value uint8) UID64 {
	cleared := u.metadata() &^ s.mask
	newMetadata := cleared | ((value << s.shift) & s.mask)
	return u.withMetadata(newMetadata)
}

func (u UID64) metadata() uint8 {
	return uint8(u >> MetadataShift)
}

func (u UID64) withMetadata(metadata uint8) UID64 {
	cleared := uint64(u) &^ (uint64(MetadataMask) << MetadataShift)
	return UID64(cleared | (uint64(metadata) << MetadataShift))
}
