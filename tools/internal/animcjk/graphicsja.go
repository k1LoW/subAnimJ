package animcjk

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1LoW/subAnimJ/tools/internal/svgpath"
)

// graphicsJaEntry is the on-disk JSONL shape of a single character entry.
type graphicsJaEntry struct {
	Character string        `json:"character"`
	Strokes   []string      `json:"strokes"`
	Medians   [][][]float64 `json:"medians"`
}

// LoadFromUpstream reads upstream graphicsJa.txt for the given character and
// builds a Glyph with strokes and medians in y-up coordinates.
func LoadFromUpstream(upstreamRoot, character string) (*Glyph, error) {
	graphicsPath := filepath.Join(upstreamRoot, "graphicsJa.txt")
	f, err := os.Open(graphicsPath)
	if err != nil {
		return nil, fmt.Errorf("open graphicsJa.txt: %w", err)
	}
	defer f.Close()

	needle := `"character":"` + character + `"`
	br := bufio.NewReaderSize(f, 1<<20)
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			if strings.Contains(line, needle) {
				return parseEntryLine(strings.TrimRight(line, "\n"), character)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read graphicsJa.txt: %w", err)
		}
	}
	return nil, fmt.Errorf("character %q not found in graphicsJa.txt", character)
}

func parseEntryLine(line, character string) (*Glyph, error) {
	var e graphicsJaEntry
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		return nil, fmt.Errorf("decode entry: %w", err)
	}
	if e.Character != character {
		return nil, fmt.Errorf("matched line had character %q, want %q", e.Character, character)
	}
	if len(e.Strokes) != len(e.Medians) {
		return nil, fmt.Errorf("strokes/medians length mismatch: %d vs %d", len(e.Strokes), len(e.Medians))
	}

	cp, err := singleRuneCodepoint(character)
	if err != nil {
		return nil, err
	}

	g := &Glyph{Character: character, Codepoint: cp, Strokes: make([]Stroke, len(e.Strokes))}
	for i, d := range e.Strokes {
		path, err := svgpath.Parse(d)
		if err != nil {
			return nil, fmt.Errorf("parse stroke %d: %w", i+1, err)
		}
		median := make([]svgpath.Point, 0, len(e.Medians[i]))
		for _, pt := range e.Medians[i] {
			if len(pt) != 2 {
				return nil, fmt.Errorf("median point %d in stroke %d not 2-tuple", len(median), i+1)
			}
			median = append(median, svgpath.Point{X: pt[0], Y: pt[1]})
		}
		g.Strokes[i] = Stroke{Brush: path, Median: median}
	}
	return g, nil
}

// WriteGraphicsJa writes/updates the entry for g in outRoot/graphicsJa.txt.
// The output file holds only the subset of modified entries.
func WriteGraphicsJa(outRoot string, g *Glyph) error {
	outPath := filepath.Join(outRoot, "graphicsJa.txt")
	existing, err := readEntries(outPath)
	if err != nil {
		return err
	}

	newLine, err := serializeEntry(g)
	if err != nil {
		return err
	}

	replaced := false
	out := make([]string, 0, len(existing)+1)
	for _, line := range existing {
		if entryCharacter(line) == g.Character {
			out = append(out, newLine)
			replaced = true
		} else {
			out = append(out, line)
		}
	}
	if !replaced {
		out = append(out, newLine)
	}

	if err := os.MkdirAll(outRoot, 0o755); err != nil {
		return err
	}
	body := strings.Join(out, "\n") + "\n"
	return os.WriteFile(outPath, []byte(body), 0o644)
}

func readEntries(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(b), "\n") {
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func entryCharacter(line string) string {
	const prefix = `{"character":"`
	if !strings.HasPrefix(line, prefix) {
		return ""
	}
	rest := line[len(prefix):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func serializeEntry(g *Glyph) (string, error) {
	e := graphicsJaEntry{
		Character: g.Character,
		Strokes:   make([]string, len(g.Strokes)),
		Medians:   make([][][]float64, len(g.Strokes)),
	}
	for i, s := range g.Strokes {
		e.Strokes[i] = s.Brush.ToGraphicsJa()
		pts := make([][]float64, len(s.Median))
		for j, p := range s.Median {
			pts[j] = []float64{roundFloat(p.X), roundFloat(p.Y)}
		}
		e.Medians[i] = pts
	}
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func roundFloat(v float64) float64 {
	if v >= 0 {
		return float64(int64(v + 0.5))
	}
	return float64(int64(v - 0.5))
}

func singleRuneCodepoint(s string) (int, error) {
	rs := []rune(s)
	if len(rs) != 1 {
		return 0, fmt.Errorf("expected single rune, got %q", s)
	}
	return int(rs[0]), nil
}
