// Package runner scans a targets directory of per-kanji JSONL files and
// applies each file's operations to the corresponding glyph, writing the
// modified svgsJa SVG and graphicsJa entry once per kanji.
package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/k1LoW/subAnimJ/tools/internal/animcjk"
	"github.com/k1LoW/subAnimJ/tools/internal/extend"
)

type Op struct {
	Op        string `json:"op"`
	Stroke    int    `json:"stroke,omitempty"`
	Direction string `json:"direction,omitempty"`
}

func RunDir(upstreamRoot, outRoot, targetsDir string) error {
	entries, err := os.ReadDir(targetsDir)
	if err != nil {
		return fmt.Errorf("read targets dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	type built struct {
		Char     string
		Filename string
	}
	var builts []built
	for _, name := range names {
		path := filepath.Join(targetsDir, name)
		kanji, codepoint, err := runFile(upstreamRoot, outRoot, path)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		builts = append(builts, built{Char: kanji, Filename: fmt.Sprintf("svgsJa/%d.svg", codepoint)})
		fmt.Fprintf(os.Stderr, "built: %s\n", name)
	}

	pairs := make([][2]string, len(builts))
	for i, b := range builts {
		pairs[i] = [2]string{b.Char, b.Filename}
	}
	if err := writePreview(outRoot, pairs); err != nil {
		return fmt.Errorf("write preview.html: %w", err)
	}
	return nil
}

func runFile(upstreamRoot, outRoot, path string) (string, int, error) {
	base := filepath.Base(path)
	kanji := strings.TrimSuffix(base, filepath.Ext(base))
	if utf8.RuneCountInString(kanji) != 1 {
		return "", 0, fmt.Errorf("filename must contain exactly one rune (got %q)", base)
	}

	ops, err := readOps(path)
	if err != nil {
		return "", 0, err
	}
	if len(ops) == 0 {
		return "", 0, fmt.Errorf("no operations")
	}

	g, err := animcjk.LoadFromUpstream(upstreamRoot, kanji)
	if err != nil {
		return "", 0, err
	}
	for i, op := range ops {
		if err := apply(g, op); err != nil {
			return "", 0, fmt.Errorf("op #%d (%+v): %w", i+1, op, err)
		}
	}

	if err := animcjk.WriteSVG(upstreamRoot, outRoot, g); err != nil {
		return "", 0, err
	}
	if err := animcjk.WriteGraphicsJa(outRoot, g); err != nil {
		return "", 0, err
	}
	return kanji, g.Codepoint, nil
}

func readOps(path string) ([]Op, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ops []Op
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		var op Op
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			return nil, fmt.Errorf("decode %q: %w", line, err)
		}
		ops = append(ops, op)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return ops, nil
}

func apply(g *animcjk.Glyph, op Op) error {
	switch op.Op {
	case "extend":
		dir, err := extend.ParseDirection(op.Direction)
		if err != nil {
			return err
		}
		return extend.Apply(g, op.Stroke, dir)
	default:
		return fmt.Errorf("unknown op %q", op.Op)
	}
}
