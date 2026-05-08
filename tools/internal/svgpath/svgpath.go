// Package svgpath parses and serializes SVG path "d" attributes restricted
// to absolute M, L, Q, C and Z commands as used by animCJK files.
package svgpath

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Point struct {
	X, Y float64
}

func (p Point) Sub(q Point) Point { return Point{p.X - q.X, p.Y - q.Y} }
func (p Point) Add(q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }
func (p Point) Scale(s float64) Point {
	return Point{p.X * s, p.Y * s}
}
func (p Point) Len() float64 {
	return math.Hypot(p.X, p.Y)
}
func (p Point) Norm() Point {
	l := p.Len()
	if l == 0 {
		return Point{}
	}
	return Point{p.X / l, p.Y / l}
}
func Distance(a, b Point) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

type Command struct {
	Op     byte
	Points []Point
}

type Path struct {
	Commands []Command
}

func numPoints(op byte) int {
	switch op {
	case 'M', 'L':
		return 1
	case 'Q':
		return 2
	case 'C':
		return 3
	case 'Z':
		return 0
	default:
		return -1
	}
}

// Parse parses a d attribute. Lowercase (relative) commands are not supported
// because animCJK normalizes everything to uppercase absolute commands.
func Parse(d string) (Path, error) {
	sc := &scanner{s: d}
	var p Path
	var lastOp byte
	for {
		sc.skipSep()
		if sc.eof() {
			break
		}
		c := sc.peek()
		var op byte
		if isAlpha(c) {
			if !isUpperOp(c) && c != 'Z' && c != 'z' {
				return p, fmt.Errorf("relative command %q not supported", c)
			}
			op = c
			if op == 'z' {
				op = 'Z'
			}
			sc.pos++
		} else {
			if lastOp == 0 {
				return p, fmt.Errorf("expected command at offset %d", sc.pos)
			}
			// Implicit repeat: after M, additional coordinates are L.
			if lastOp == 'M' {
				op = 'L'
			} else {
				op = lastOp
			}
		}
		n := numPoints(op)
		if n < 0 {
			return p, fmt.Errorf("unsupported command %q", op)
		}
		cmd := Command{Op: op}
		for i := 0; i < n; i++ {
			x, ok := sc.readNumber()
			if !ok {
				return p, fmt.Errorf("expected number at offset %d", sc.pos)
			}
			y, ok := sc.readNumber()
			if !ok {
				return p, fmt.Errorf("expected number at offset %d", sc.pos)
			}
			cmd.Points = append(cmd.Points, Point{X: x, Y: y})
		}
		p.Commands = append(p.Commands, cmd)
		lastOp = op
	}
	return p, nil
}

// ToSVG serializes the path using space-separated integer coordinates,
// matching the style of svgsJa/*.svg files.
func (p Path) ToSVG() string {
	return p.serialize(" ")
}

// ToGraphicsJa serializes the path using comma between x and y of each point,
// matching the style of graphicsJa.txt entries.
func (p Path) ToGraphicsJa() string {
	return p.serialize(",")
}

func (p Path) serialize(coordSep string) string {
	var b strings.Builder
	for _, cmd := range p.Commands {
		b.WriteByte(cmd.Op)
		for i, pt := range cmd.Points {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(formatInt(pt.X))
			b.WriteString(coordSep)
			b.WriteString(formatInt(pt.Y))
		}
	}
	return b.String()
}

func formatInt(v float64) string {
	return strconv.FormatInt(int64(math.Round(v)), 10)
}

// Last returns the destination (anchor) point of a command. Z returns the
// implicit closing point which the caller must track separately.
func (c Command) Last() (Point, bool) {
	if len(c.Points) == 0 {
		return Point{}, false
	}
	return c.Points[len(c.Points)-1], true
}

type scanner struct {
	s   string
	pos int
}

func (sc *scanner) eof() bool { return sc.pos >= len(sc.s) }
func (sc *scanner) peek() byte {
	if sc.eof() {
		return 0
	}
	return sc.s[sc.pos]
}
func (sc *scanner) skipSep() {
	for !sc.eof() {
		c := sc.s[sc.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' {
			sc.pos++
		} else {
			break
		}
	}
}

func (sc *scanner) readNumber() (float64, bool) {
	sc.skipSep()
	start := sc.pos
	if !sc.eof() && (sc.s[sc.pos] == '-' || sc.s[sc.pos] == '+') {
		sc.pos++
	}
	sawDigit := false
	for !sc.eof() {
		c := sc.s[sc.pos]
		switch {
		case c >= '0' && c <= '9':
			sc.pos++
			sawDigit = true
		case c == '.':
			sc.pos++
		case (c == 'e' || c == 'E') && sawDigit:
			sc.pos++
			if !sc.eof() && (sc.s[sc.pos] == '-' || sc.s[sc.pos] == '+') {
				sc.pos++
			}
		default:
			goto done
		}
	}
done:
	if sc.pos == start || !sawDigit {
		return 0, false
	}
	v, err := strconv.ParseFloat(sc.s[start:sc.pos], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func isAlpha(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}
func isUpperOp(c byte) bool {
	switch c {
	case 'M', 'L', 'Q', 'C', 'Z':
		return true
	}
	return false
}
