// Package parts loads radical SVG parts authored under parts/.
//
// A part SVG is laid out like animCJK upstream files but coordinates are
// y-down within viewBox 0..1024:
//
//   <path id="d{N}" d="..."/>           (N brush outline polygons)
//   <clipPath id="_clip{N}">...</clipPath>
//   <g clip-path="url(#_clip{N})">
//     <path d="M ... L ..."/>           (median polyline)
//   </g>
//
// Strokes are returned in y-up coordinates so they line up with
// graphicsJa convention used elsewhere in the codebase.
package parts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/k1LoW/subAnimJ/tools/internal/animcjk"
	"github.com/k1LoW/subAnimJ/tools/internal/svgpath"
)

type Part struct {
	Name    string
	Strokes []animcjk.Stroke
}

var (
	brushPathRe  = regexp.MustCompile(`<path[^>]*\bid="d(\d+)"[^>]*\bd="([^"]+)"`)
	clipMedianRe = regexp.MustCompile(`clip-path="url\(#_clip(\d+)\)"[^>]*>\s*<path[^>]*\bd="([^"]+)"`)
)

func Load(path string) (*Part, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read part %s: %w", path, err)
	}
	text := string(src)

	brushes, err := collectBrushes(text)
	if err != nil {
		return nil, err
	}
	medians, err := collectMedians(text)
	if err != nil {
		return nil, err
	}
	if len(brushes) == 0 {
		return nil, fmt.Errorf("part %s: no <path id=\"d{N}\"> brushes found", path)
	}
	if len(brushes) != len(medians) {
		return nil, fmt.Errorf("part %s: brush count %d != median count %d", path, len(brushes), len(medians))
	}

	indices := make([]int, 0, len(brushes))
	for n := range brushes {
		indices = append(indices, n)
	}
	sort.Ints(indices)

	strokes := make([]animcjk.Stroke, 0, len(indices))
	for _, n := range indices {
		mp, ok := medians[n]
		if !ok {
			return nil, fmt.Errorf("part %s: brush d%d has no matching median", path, n)
		}
		strokes = append(strokes, animcjk.Stroke{
			Brush:  animcjk.FlipPathY(brushes[n]),
			Median: flipPoints(medianAnchors(mp)),
		})
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return &Part{Name: name, Strokes: strokes}, nil
}

func collectBrushes(text string) (map[int]svgpath.Path, error) {
	out := map[int]svgpath.Path{}
	for _, m := range brushPathRe.FindAllStringSubmatch(text, -1) {
		n, _ := strconv.Atoi(m[1])
		p, err := svgpath.Parse(m[2])
		if err != nil {
			return nil, fmt.Errorf("brush d%d: %w", n, err)
		}
		out[n] = p
	}
	return out, nil
}

func collectMedians(text string) (map[int]svgpath.Path, error) {
	out := map[int]svgpath.Path{}
	for _, m := range clipMedianRe.FindAllStringSubmatch(text, -1) {
		n, _ := strconv.Atoi(m[1])
		p, err := svgpath.Parse(m[2])
		if err != nil {
			return nil, fmt.Errorf("median %d: %w", n, err)
		}
		out[n] = p
	}
	return out, nil
}

func medianAnchors(p svgpath.Path) []svgpath.Point {
	var pts []svgpath.Point
	for _, c := range p.Commands {
		if c.Op == 'M' || c.Op == 'L' {
			if len(c.Points) > 0 {
				pts = append(pts, c.Points[0])
			}
		}
	}
	return pts
}

func flipPoints(pts []svgpath.Point) []svgpath.Point {
	out := make([]svgpath.Point, len(pts))
	for i, p := range pts {
		out[i] = animcjk.FlipY(p)
	}
	return out
}
