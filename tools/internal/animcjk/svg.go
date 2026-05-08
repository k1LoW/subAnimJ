package animcjk

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/k1LoW/subAnimJ/tools/internal/svgpath"
)

// WriteSVG copies the upstream SVG template for the glyph and rewrites the
// d attribute of every brush polygon and median path to reflect g.Strokes.
func WriteSVG(upstreamRoot, outRoot string, g *Glyph) error {
	srcPath := filepath.Join(upstreamRoot, "svgsJa", strconv.Itoa(g.Codepoint)+".svg")
	srcBytes, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read upstream svg %s: %w", srcPath, err)
	}
	src := string(srcBytes)

	for i, s := range g.Strokes {
		strokeNo := i + 1
		brushSVG := FlipPathY(s.Brush).ToSVG()
		brushID := fmt.Sprintf("z%dd%d", g.Codepoint, strokeNo)
		var replaceErr error
		src, replaceErr = replaceDByID(src, brushID, brushSVG)
		if replaceErr != nil {
			return fmt.Errorf("update brush stroke %d: %w", strokeNo, replaceErr)
		}

		clipRef := fmt.Sprintf("z%dc%d", g.Codepoint, strokeNo)
		medianSVG := medianToSVGPath(s.Median)
		src, replaceErr = replaceDByClipPath(src, clipRef, medianSVG)
		if replaceErr != nil {
			return fmt.Errorf("update median stroke %d: %w", strokeNo, replaceErr)
		}
	}

	outDir := filepath.Join(outRoot, "svgsJa")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	outPath := filepath.Join(outDir, strconv.Itoa(g.Codepoint)+".svg")
	return os.WriteFile(outPath, []byte(src), 0o644)
}

// medianToSVGPath serializes a y-up polyline into a y-down "M ... L ... L ..." string.
func medianToSVGPath(pts []svgpath.Point) string {
	if len(pts) == 0 {
		return ""
	}
	var b strings.Builder
	for i, p := range pts {
		if i == 0 {
			b.WriteByte('M')
		} else {
			b.WriteByte('L')
		}
		b.WriteString(strconv.FormatInt(roundInt(p.X), 10))
		b.WriteByte(' ')
		b.WriteString(strconv.FormatInt(roundInt(YFlipBase-p.Y), 10))
	}
	return b.String()
}

func roundInt(v float64) int64 {
	return int64(math.Round(v))
}

// replaceDByID finds the <path id="<id>" ... d="..."/> tag and replaces the d value.
func replaceDByID(src, id, newD string) (string, error) {
	re := regexp.MustCompile(`<path[^>]*\bid="` + regexp.QuoteMeta(id) + `"[^>]*>`)
	loc := re.FindStringIndex(src)
	if loc == nil {
		return src, fmt.Errorf("path with id=%q not found", id)
	}
	tag := src[loc[0]:loc[1]]
	newTag, err := replaceDInTag(tag, newD)
	if err != nil {
		return src, err
	}
	return src[:loc[0]] + newTag + src[loc[1]:], nil
}

// replaceDByClipPath finds the <path ... clip-path="url(#<ref>)" ... d="..."/> tag and replaces the d value.
func replaceDByClipPath(src, ref, newD string) (string, error) {
	re := regexp.MustCompile(`<path[^>]*\bclip-path="url\(#` + regexp.QuoteMeta(ref) + `\)"[^>]*>`)
	loc := re.FindStringIndex(src)
	if loc == nil {
		return src, fmt.Errorf("path with clip-path url(#%s) not found", ref)
	}
	tag := src[loc[0]:loc[1]]
	newTag, err := replaceDInTag(tag, newD)
	if err != nil {
		return src, err
	}
	return src[:loc[0]] + newTag + src[loc[1]:], nil
}

func replaceDInTag(tag, newD string) (string, error) {
	idx := strings.Index(tag, ` d="`)
	if idx < 0 {
		return "", fmt.Errorf("d attribute not found in tag")
	}
	start := idx + len(` d="`)
	end := strings.Index(tag[start:], `"`)
	if end < 0 {
		return "", fmt.Errorf("unterminated d attribute")
	}
	end += start
	return tag[:start] + newD + tag[end:], nil
}
