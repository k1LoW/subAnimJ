package extend

import (
	"math"
	"testing"

	"github.com/k1LoW/subAnimJ/tools/internal/animcjk"
)

const upstreamRoot = "../../../vendor/animCJK"

func TestApply_HiStroke3(t *testing.T) {
	g, err := animcjk.LoadFromUpstream(upstreamRoot, "日")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	origLast := g.Strokes[2].Median[len(g.Strokes[2].Median)-1]
	if math.Abs(origLast.X-586) > 0.5 || math.Abs(origLast.Y-431) > 0.5 {
		t.Fatalf("unexpected upstream median end %v", origLast)
	}

	if err := Apply(g, 3, Horizontal, BothSides); err != nil {
		t.Fatalf("apply: %v", err)
	}

	newLast := g.Strokes[2].Median[len(g.Strokes[2].Median)-1]
	if newLast.X <= origLast.X {
		t.Fatalf("median end did not move right: %v -> %v", origLast, newLast)
	}
	if newLast.X < 690 || newLast.X > 700 {
		t.Fatalf("median end x should reach right vertical (~693), got %v", newLast)
	}

	// The other end (left side) of stroke 3 must be unchanged.
	first := g.Strokes[2].Median[0]
	if math.Abs(first.X-347) > 0.5 || math.Abs(first.Y-407) > 0.5 {
		t.Fatalf("left end of median should be unchanged, got %v", first)
	}
}

func TestApply_DirectionMismatch(t *testing.T) {
	g, err := animcjk.LoadFromUpstream(upstreamRoot, "日")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := Apply(g, 3, Vertical, BothSides); err == nil {
		t.Fatal("expected error for vertical direction on horizontal stroke")
	}
}

func TestApply_EndOnly(t *testing.T) {
	g, err := animcjk.LoadFromUpstream(upstreamRoot, "日")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	origFirst := g.Strokes[2].Median[0]
	origLast := g.Strokes[2].Median[len(g.Strokes[2].Median)-1]
	if err := Apply(g, 3, Horizontal, EndOnly); err != nil {
		t.Fatalf("apply: %v", err)
	}
	first := g.Strokes[2].Median[0]
	if first != origFirst {
		t.Fatalf("EndOnly should not move start: %v -> %v", origFirst, first)
	}
	last := g.Strokes[2].Median[len(g.Strokes[2].Median)-1]
	if last == origLast {
		t.Fatalf("EndOnly should move end")
	}
}
