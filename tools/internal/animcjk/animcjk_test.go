package animcjk

import (
	"strings"
	"testing"
)

const upstreamRoot = "../../../vendor/animCJK"

func TestLoadAndSerialize_Hi(t *testing.T) {
	g, err := LoadFromUpstream(upstreamRoot, "日")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if g.Codepoint != 0x65E5 {
		t.Fatalf("codepoint: got %x want 65E5", g.Codepoint)
	}
	if len(g.Strokes) != 4 {
		t.Fatalf("expected 4 strokes, got %d", len(g.Strokes))
	}
	out, err := serializeEntry(g)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	// Roundtrip check: every original strokes/medians substring should still appear.
	if !strings.Contains(out, `"character":"日"`) {
		t.Fatalf("missing character key: %s", out[:200])
	}
	expectedMedian := `[[279,689],[319,651],[329,603],[305,66]]`
	if !strings.Contains(out, expectedMedian) {
		t.Fatalf("median[0] mismatch; out=\n%s", out)
	}
	expectedStrokeStart := `M349,664C334,682`
	if !strings.Contains(out, expectedStrokeStart) {
		t.Fatalf("stroke[0] start mismatch; out=\n%s", out)
	}
}
