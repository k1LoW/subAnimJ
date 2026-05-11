// Package compose replaces a contiguous run of strokes in a glyph with the
// strokes of a Part, scaling them to fit the bounding box of the strokes
// being replaced.
//
// The number of strokes in the part must match the run length, so that
// stroke ids in the output SVG (z<cp>d{N}, z<cp>c{N}) keep their
// numbering and the rest of the glyph (later strokes) is unaffected.
package compose

import (
	"fmt"
	"math"

	"github.com/k1LoW/subAnimJ/tools/internal/animcjk"
	"github.com/k1LoW/subAnimJ/tools/internal/parts"
	"github.com/k1LoW/subAnimJ/tools/internal/svgpath"
)

type bbox struct {
	minX, maxX, minY, maxY float64
}

func (b bbox) width() float64  { return b.maxX - b.minX }
func (b bbox) height() float64 { return b.maxY - b.minY }

// Apply replaces glyph strokes [start, start+len(part.Strokes)-1]
// (1-origin) with the part's strokes, transformed so the part's
// fitStroke (1-origin within the part) lines up with the corresponding
// stroke in the target glyph. The same affine transform is then applied
// to every other stroke in the part, preserving the part's relative
// proportions. start defaults to 1; fitStroke <= 0 means use the bbox of
// the entire part. offsetX/offsetY (in graphicsJa y-up coordinates,
// -x = left, +y = up) are applied to every transformed point after the
// fit transform, providing a knob for fine-tuning placement when the
// fit-derived position is slightly off.
func Apply(g *animcjk.Glyph, part *parts.Part, start, fitStroke int, offsetX, offsetY float64) error {
	if start <= 0 {
		start = 1
	}
	n := len(part.Strokes)
	if n == 0 {
		return fmt.Errorf("part %q has no strokes", part.Name)
	}
	end := start + n - 1
	if end > len(g.Strokes) {
		return fmt.Errorf("part %q (%d strokes from %d) extends past glyph stroke count %d", part.Name, n, start, len(g.Strokes))
	}

	var sourceStrokes, targetStrokes []animcjk.Stroke
	if fitStroke <= 0 {
		sourceStrokes = part.Strokes
		targetStrokes = g.Strokes[start-1 : start-1+n]
	} else {
		if fitStroke > n {
			return fmt.Errorf("fit_stroke %d out of range (1..%d)", fitStroke, n)
		}
		sourceStrokes = []animcjk.Stroke{part.Strokes[fitStroke-1]}
		targetStrokes = []animcjk.Stroke{g.Strokes[start-1+fitStroke-1]}
	}

	source := strokesBBox(sourceStrokes)
	target := strokesBBox(targetStrokes)
	if source.width() == 0 || source.height() == 0 {
		return fmt.Errorf("part %q fit stroke has zero-width or zero-height bounding box", part.Name)
	}
	if target.width() == 0 || target.height() == 0 {
		return fmt.Errorf("target fit stroke has zero-width or zero-height bounding box")
	}
	sx := target.width() / source.width()
	sy := target.height() / source.height()
	tx := target.minX - source.minX*sx
	ty := target.minY - source.minY*sy
	xform := func(p svgpath.Point) svgpath.Point {
		return svgpath.Point{X: p.X*sx + tx + offsetX, Y: p.Y*sy + ty + offsetY}
	}

	for i, ps := range part.Strokes {
		g.Strokes[start-1+i] = animcjk.Stroke{
			Brush:  transformPath(ps.Brush, xform),
			Median: transformPoints(ps.Median, xform),
		}
	}
	return nil
}

func strokesBBox(strokes []animcjk.Stroke) bbox {
	b := bbox{minX: math.Inf(1), maxX: math.Inf(-1), minY: math.Inf(1), maxY: math.Inf(-1)}
	for _, s := range strokes {
		for _, c := range s.Brush.Commands {
			for _, p := range c.Points {
				b = expand(b, p)
			}
		}
		for _, p := range s.Median {
			b = expand(b, p)
		}
	}
	return b
}

func expand(b bbox, p svgpath.Point) bbox {
	if p.X < b.minX {
		b.minX = p.X
	}
	if p.X > b.maxX {
		b.maxX = p.X
	}
	if p.Y < b.minY {
		b.minY = p.Y
	}
	if p.Y > b.maxY {
		b.maxY = p.Y
	}
	return b
}

func transformPath(p svgpath.Path, fn func(svgpath.Point) svgpath.Point) svgpath.Path {
	out := svgpath.Path{Commands: make([]svgpath.Command, len(p.Commands))}
	for i, c := range p.Commands {
		nc := svgpath.Command{Op: c.Op}
		if len(c.Points) > 0 {
			nc.Points = make([]svgpath.Point, len(c.Points))
			for j, pt := range c.Points {
				nc.Points[j] = fn(pt)
			}
		}
		out.Commands[i] = nc
	}
	return out
}

func transformPoints(pts []svgpath.Point, fn func(svgpath.Point) svgpath.Point) []svgpath.Point {
	out := make([]svgpath.Point, len(pts))
	for i, p := range pts {
		out[i] = fn(p)
	}
	return out
}
