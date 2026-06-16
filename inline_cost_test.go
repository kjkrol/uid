package uid

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// inlineBudgets caps the compiler's inline cost (AST-node count) for the hot
// bit-twiddling helpers. These functions are inlined into callers' tight loops
// (entity lookups, validity checks), so their cost is added to the caller's own
// inline budget — keeping them small lets callers stay under the inliner's hard
// limit of 80. Values are the current cost plus a small margin; tighten or relax
// deliberately, never to silence a regression you have not understood.
var inlineBudgets = map[string]int{
	"UID64.Unpack":     16,
	"UID64.Index":      10,
	"UID64.Generation": 12,
	"UID64.metadata":   10,
}

// TestInlineCost guards the inline budget of the helpers in inlineBudgets.
//
// Inline cost is a compile-time property, not a runtime one, so it cannot be
// measured with a benchmark. Instead we recompile the package with -gcflags=-m=2
// and parse the "can inline F with cost N" diagnostics the compiler emits.
// A fresh GOCACHE forces a cold build so the diagnostics are actually produced
// (a cached build emits nothing).
func TestInlineCost(t *testing.T) {
	cmd := exec.Command("go", "build", "-gcflags=-m=2", ".")
	cmd.Env = append(os.Environ(), "GOCACHE="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile failed: %v\n%s", err, out)
	}

	re := regexp.MustCompile(`can inline (\S+) with cost (\d+)`)
	cost := make(map[string]int)
	for line := range strings.SplitSeq(string(out), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			c, _ := strconv.Atoi(m[2])
			cost[m[1]] = c
		}
	}

	for fn, budget := range inlineBudgets {
		switch c, ok := cost[fn]; {
		case !ok:
			t.Errorf("%s: compiler did not report it as inlinable — it may have stopped inlining entirely", fn)
		case c > budget:
			t.Errorf("%s: inline cost %d exceeds budget %d", fn, c, budget)
		default:
			t.Logf("%s: inline cost %d (budget %d)", fn, c, budget)
		}
	}
}
