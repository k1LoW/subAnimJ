// Package animcjk reads and writes animCJK svgsJa SVG files and graphicsJa.txt
// JSONL entries. Internally all coordinates use the y-up convention from
// graphicsJa (y_svg = YFlipBase - y).
package animcjk

import "github.com/k1LoW/subAnimJ/tools/internal/svgpath"

// YFlipBase converts between graphicsJa (y-up) coordinates and SVG (y-down)
// coordinates. Verified empirically against animCJK upstream files.
const YFlipBase = 900.0

type Stroke struct {
	Brush  svgpath.Path     // brush outline polygon, y-up
	Median []svgpath.Point  // centerline polyline, y-up
}

type Glyph struct {
	Character string
	Codepoint int
	Strokes   []Stroke
}

// FlipY returns p mirrored across the y axis at YFlipBase.
func FlipY(p svgpath.Point) svgpath.Point {
	return svgpath.Point{X: p.X, Y: YFlipBase - p.Y}
}

// FlipPathY returns a copy of path with every point's y mirrored.
func FlipPathY(path svgpath.Path) svgpath.Path {
	out := svgpath.Path{Commands: make([]svgpath.Command, len(path.Commands))}
	for i, c := range path.Commands {
		nc := svgpath.Command{Op: c.Op}
		if len(c.Points) > 0 {
			nc.Points = make([]svgpath.Point, len(c.Points))
			for j, p := range c.Points {
				nc.Points[j] = FlipY(p)
			}
		}
		out.Commands[i] = nc
	}
	return out
}
