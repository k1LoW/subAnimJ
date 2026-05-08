// Package extend implements angle-preserving stroke extension on a Glyph.
package extend

import (
	"fmt"
	"math"

	"github.com/k1LoW/subAnimJ/tools/internal/animcjk"
	"github.com/k1LoW/subAnimJ/tools/internal/svgpath"
)

type Direction int

const (
	Horizontal Direction = iota
	Vertical
)

func ParseDirection(s string) (Direction, error) {
	switch s {
	case "horizontal", "h":
		return Horizontal, nil
	case "vertical", "v":
		return Vertical, nil
	}
	return 0, fmt.Errorf("invalid direction %q (want horizontal|vertical|h|v)", s)
}

// Side selects which median end(s) of the target stroke are eligible for
// extension. "both" (default) extends any end whose tangent matches the
// requested direction and whose gap to another stroke exceeds the threshold.
// "start" / "end" restricts extension to that end only, even when the other
// end would also qualify.
type Side int

const (
	BothSides Side = iota
	StartOnly
	EndOnly
)

func ParseSide(s string) (Side, error) {
	switch s {
	case "", "both":
		return BothSides, nil
	case "start":
		return StartOnly, nil
	case "end":
		return EndOnly, nil
	}
	return 0, fmt.Errorf("invalid side %q (want both|start|end)", s)
}

// minExtensionGap is the threshold below which an end is considered
// "already connected" and not extended. Empirically tuned: 日 stroke 3's
// left end (gap ~27 to the left vertical) should NOT be extended (the brush
// outlines overlap enough), but 町 stroke 4 (gap ~54) SHOULD be extended.
const minExtensionGap = 30.0

// Apply extends the eligible end(s) of strokeNum (1-origin) of g along its
// tangent. Each end is considered independently: if the end is allowed by
// side AND its tangent matches dir AND there is a nearer-stroke median within
// ray range AND the gap exceeds the threshold, the end is translated by
// tangent * d. Brush polygon vertices on the same half (closer to that end
// than to the opposite end) are translated by the same vector, preserving
// the stroke angle.
func Apply(g *animcjk.Glyph, strokeNum int, dir Direction, side Side) error {
	if strokeNum < 1 || strokeNum > len(g.Strokes) {
		return fmt.Errorf("stroke index %d out of range (1..%d)", strokeNum, len(g.Strokes))
	}
	idx := strokeNum - 1
	s := &g.Strokes[idx]
	if len(s.Median) < 2 {
		return fmt.Errorf("stroke %d median has fewer than 2 points", strokeNum)
	}

	last := len(s.Median) - 1
	startPt := s.Median[0]
	endPt := s.Median[last]
	startTangent := startPt.Sub(s.Median[1]).Norm()
	endTangent := endPt.Sub(s.Median[last-1]).Norm()

	startAllowed := side == BothSides || side == StartOnly
	endAllowed := side == BothSides || side == EndOnly

	var startDelta, endDelta svgpath.Point
	var startOK, endOK bool
	if startAllowed {
		startDelta, startOK = tryExtendEnd(g, idx, startPt, startTangent, dir)
	}
	if endAllowed {
		endDelta, endOK = tryExtendEnd(g, idx, endPt, endTangent, dir)
	}

	if !startOK && !endOK {
		return fmt.Errorf("stroke %d: no end could be extended in direction %s side %s", strokeNum, dirString(dir), sideString(side))
	}

	if startOK {
		s.Median[0] = startPt.Add(startDelta)
	}
	if endOK {
		s.Median[last] = endPt.Add(endDelta)
	}

	s.Brush = translateBrushHalves(s.Brush, startPt, endPt, startDelta, endDelta, startOK, endOK)
	return nil
}

func tryExtendEnd(g *animcjk.Glyph, idx int, p, tangent svgpath.Point, dir Direction) (svgpath.Point, bool) {
	if !directionMatches(tangent, dir) {
		return svgpath.Point{}, false
	}
	d, ok := nearestRayHit(g, idx, p, tangent)
	if !ok {
		return svgpath.Point{}, false
	}
	if d <= minExtensionGap {
		return svgpath.Point{}, false
	}
	return svgpath.Point{X: tangent.X * d, Y: tangent.Y * d}, true
}

func dirString(d Direction) string {
	if d == Horizontal {
		return "horizontal"
	}
	return "vertical"
}

func sideString(s Side) string {
	switch s {
	case StartOnly:
		return "start"
	case EndOnly:
		return "end"
	default:
		return "both"
	}
}

func directionMatches(tangent svgpath.Point, dir Direction) bool {
	switch dir {
	case Horizontal:
		return math.Abs(tangent.X) > math.Abs(tangent.Y)
	case Vertical:
		return math.Abs(tangent.Y) > math.Abs(tangent.X)
	}
	return false
}

// nearestRayHit returns the smallest positive distance s such that the ray
// (p0 + s*t) hits a segment of any other stroke's median.
func nearestRayHit(g *animcjk.Glyph, idx int, p0, t svgpath.Point) (float64, bool) {
	const eps = 1e-6
	minS := math.Inf(1)
	found := false
	for j, other := range g.Strokes {
		if j == idx {
			continue
		}
		for k := 0; k+1 < len(other.Median); k++ {
			a := other.Median[k]
			b := other.Median[k+1]
			s, ok := raySegment(p0, t, a, b)
			if !ok {
				continue
			}
			if s < eps {
				continue
			}
			if s < minS {
				minS = s
				found = true
			}
		}
	}
	return minS, found
}

func raySegment(p0, t, a, b svgpath.Point) (float64, bool) {
	dx := b.X - a.X
	dy := b.Y - a.Y
	det := t.Y*dx - t.X*dy
	if math.Abs(det) < 1e-9 {
		return 0, false
	}
	apx := a.X - p0.X
	apy := a.Y - p0.Y
	s := (-apx*dy + apy*dx) / det
	u := (t.X*apy - t.Y*apx) / det
	if s <= 0 || u < -1e-6 || u > 1+1e-6 {
		return 0, false
	}
	return s, true
}

// translateBrushHalves splits brush vertices into "start half" (closer to
// startPt than endPt) and "end half" (closer to endPt). Vertices in each
// half receive the corresponding delta. If a half is not being extended
// (its OK flag is false) the delta is zero and vertices stay put.
func translateBrushHalves(brush svgpath.Path, startPt, endPt, startDelta, endDelta svgpath.Point, startOK, endOK bool) svgpath.Path {
	out := svgpath.Path{Commands: make([]svgpath.Command, len(brush.Commands))}
	for i, cmd := range brush.Commands {
		nc := svgpath.Command{Op: cmd.Op}
		if len(cmd.Points) > 0 {
			nc.Points = make([]svgpath.Point, len(cmd.Points))
			for j, p := range cmd.Points {
				nearStart := svgpath.Distance(p, startPt) < svgpath.Distance(p, endPt)
				switch {
				case nearStart && startOK:
					nc.Points[j] = p.Add(startDelta)
				case !nearStart && endOK:
					nc.Points[j] = p.Add(endDelta)
				default:
					nc.Points[j] = p
				}
			}
		}
		out.Commands[i] = nc
	}
	return out
}
