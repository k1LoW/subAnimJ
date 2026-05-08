package svgpath

import "testing"

func TestRoundTripSVG(t *testing.T) {
	in := "M349 236C334 218 313 204 298 203C288 202 278 203 274 211Z"
	p, err := Parse(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := p.ToSVG()
	if got != in {
		t.Fatalf("svg roundtrip mismatch:\n got %q\nwant %q", got, in)
	}
}

func TestRoundTripGraphicsJa(t *testing.T) {
	in := "M349,664C334,682 313,696 298,697C288,698 278,697 274,689Z"
	p, err := Parse(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := p.ToGraphicsJa()
	if got != in {
		t.Fatalf("graphicsJa roundtrip mismatch:\n got %q\nwant %q", got, in)
	}
}

func TestParseLineToPolyline(t *testing.T) {
	in := "M347 493L586 469"
	p, err := Parse(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(p.Commands) != 2 {
		t.Fatalf("got %d commands, want 2", len(p.Commands))
	}
}
