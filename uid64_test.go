package uid

import (
	"testing"
)

func TestUID64_Components(t *testing.T) {
	var wantIndex uint32 = 0x12345678
	// We use a mask because Generation is now 24 bits (max 0xFFFFFF)
	var wantGen uint32 = 0xABCDEF & GenerationMask

	u := New(wantGen, wantIndex)

	if u.Index() != wantIndex {
		t.Errorf("Index(): got %d, want %d", u.Index(), wantIndex)
	}
	if u.Generation() != wantGen {
		t.Errorf("Generation(): got %d, want %d", u.Generation(), wantGen)
	}
	if u.metadata() != 0 {
		t.Errorf("Metadata(): got %X, want 0", u.metadata())
	}
}

// Scenario: New generation with the same index (slot recycling)
func TestUID64_GenerationIncrement(t *testing.T) {
	index := uint32(42)
	oldGen := uint32(1)
	newGen := uint32(2)

	oldUID := New(oldGen, index)
	newUID := New(newGen, index)

	if oldUID.Index() != newUID.Index() {
		t.Errorf("UIDs should share the same index: %d == %d", oldUID.Index(), newUID.Index())
	}

	if oldUID == newUID {
		t.Error("UIDs with different generations must not be equal")
	}

	if newUID.Generation() <= oldUID.Generation() {
		t.Errorf("new UID should have higher generation: %d > %d", newUID.Generation(), oldUID.Generation())
	}
}

func TestUID64_MaxValues(t *testing.T) {
	maxIndex := uint32(IndexMask)
	maxGen := uint32(GenerationMask)

	u := New(maxGen, maxIndex)

	if u.Index() != maxIndex {
		t.Errorf("max index mismatch: got %d, want %d", u.Index(), maxIndex)
	}
	if u.Generation() != maxGen {
		t.Errorf("max generation mismatch: got %d, want %d", u.Generation(), maxGen)
	}
}

func TestUID64_GenerationOverflow(t *testing.T) {
	// We check if providing too large a generation (e.g. full 32-bits)
	// is safely truncated by the mask during creation and does not overwrite Metadata.
	overflowGen := uint32(0xFFFFFFFF)
	index := uint32(5)

	u := New(overflowGen, index)

	expectedGen := uint32(GenerationMask)
	if u.Generation() != expectedGen {
		t.Errorf("generation not masked correctly: got %X, want %X", u.Generation(), expectedGen)
	}

	if u.metadata() != 0 {
		t.Errorf("generation overflowed into metadata: got %X", u.metadata())
	}
}

func TestUID64_MetaSegment(t *testing.T) {
	// We create a test segment (e.g. 3 bits at offset 2)
	testSegment := NewMetaSegment(3, 2)
	u := New(1, 100)

	// We assign a correct value within the limit (5 = binary 101)
	u2 := u.WithMetaSegment(testSegment, 5)

	if u2.MetaSegment(testSegment) != 5 {
		t.Errorf("MetaSegment read: got %d, want 5", u2.MetaSegment(testSegment))
	}

	// We check if segment modification did not corrupt Index or Generation
	if u2.Index() != 100 || u2.Generation() != 1 {
		t.Errorf("MetaSegment mutation corrupted core identifiers")
	}

	// We check safe bit truncation (Overflow Protection).
	// Writing the value 255 (0xFF) into a 3-bit segment should silently truncate the value to the maximum (7).
	u3 := u.WithMetaSegment(testSegment, 0xFF)
	if u3.MetaSegment(testSegment) != 7 {
		t.Errorf("MetaSegment failed to truncate overflow: got %d, want 7", u3.MetaSegment(testSegment))
	}
}
