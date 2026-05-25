#!/usr/bin/env python3
"""Generate animCJK-format SVG files and graphicsJa.txt entries for Japanese
punctuation extracted from the Klee One font.

Usage (from repository root):

    python3 tools/extract_punct.py

Writes svgsJa/{codepoint}.svg for each codepoint listed in PUNCT below and
updates graphicsJa.txt with the matching JSONL entries.
"""

from __future__ import annotations

import json
import os
import re
import sys

from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.ttLib import TTFont

FONT_PATH = "vendor/Klee/fonts/ttf/KleeOne-Regular.ttf"
SVGS_DIR = "svgsJa"
GRAPHICS_FILE = "graphicsJa.txt"

UPM = 1000.0
# Map Klee's 1000-unit em to the 1024-unit viewBox 1:1, so high punctuation
# (「 ︑ ︒ ﹂) sits up against the top of the viewBox like it does on a
# typeset line. Deepest descender in the set is y_font=-85 (。), which lands
# at y_svg=987 — still inside the viewBox.
SCALE = 1024.0 / UPM  # 1.024
BASELINE_Y_SVG = 900.0
X_OFFSET = (1024.0 - UPM * SCALE) / 2.0  # 0

# Medians in Klee font coordinates (y-up). One polyline per stroke. All of
# these punctuation marks are single-stroke glyphs. Points are kept clearly
# inside the brush polygon to avoid the median escaping the clip region.
MEDIANS: dict[int, list[tuple[float, float]]] = {
    0x3001: [(140, 110), (210, 30), (260, -40)],
    0x3002: [
        (214, 162),
        (175, 156),
        (141, 138),
        (116, 110),
        (101, 76),
        (101, 33),
        (116, -10),
        (141, -38),
        (175, -57),
        (214, -64),
        (253, -57),
        (287, -38),
        (312, -10),
        (327, 33),
        (327, 76),
        (312, 110),
        (287, 138),
        (253, 156),
        (214, 162),
    ],
    0x300C: [(920, 778), (760, 778), (760, 160)],
    0x300D: [(220, 575), (220, -35), (90, -35)],
    0xFE11: [(720, 780), (805, 705), (885, 615)],
    0xFE12: [
        (784, 780),
        (745, 773),
        (711, 754),
        (686, 724),
        (673, 687),
        (673, 647),
        (686, 611),
        (711, 580),
        (745, 561),
        (784, 554),
        (823, 561),
        (857, 580),
        (882, 611),
        (895, 647),
        (895, 687),
        (882, 724),
        (857, 754),
        (823, 773),
        (784, 780),
    ],
    0xFE41: [(310, 110), (910, 110), (910, -50)],
    0xFE42: [(90, 790), (90, 625), (705, 625)],
}

SVG_HEADER = """<!--
subAnimJ
Derived from:
    Klee One - https://github.com/fontworks-fonts/Klee
    Copyright 2020 The Klee Project Authors
You can redistribute and/or modify this file under the terms of the SIL Open Font License, Version 1.1
as published by SIL International.
You should have received a copy of this license along with this file.
If not, see https://openfontlicense.org/.
-->"""

SVG_STYLE = """<style>
<![CDATA[
@keyframes zk {
\tto {
\t\tstroke-dashoffset:0;
\t}
}
svg.acjk path[clip-path] {
\t--t:0.8s;
\tanimation:zk var(--t) linear forwards var(--d);
\tstroke-dasharray:3337;
\tstroke-dashoffset:3339;
\tstroke-width:128;
\tstroke-linecap:round;
\tfill:none;
\tstroke:#000;
}
svg.acjk path[id] {fill:#ccc;}
]]>
</style>"""


def to_svg_xy(x: float, y: float) -> tuple[float, float]:
    """Klee font coords (y-up, baseline at 0) -> animCJK SVG coords (y-down, 1024 viewBox).

    The Klee baseline (y_font=0) maps to y_svg=BASELINE_Y_SVG so descenders
    fall below the baseline (towards the bottom of the viewBox) just like in
    text. x is scaled and shifted to center the em horizontally.
    """
    return x * SCALE + X_OFFSET, BASELINE_Y_SVG - y * SCALE


def to_graphics_xy(x: float, y: float) -> tuple[float, float]:
    """Klee font coords -> graphicsJa coords (y-up, y_svg + y_graphics = 900)."""
    sx, sy = to_svg_xy(x, y)
    return sx, 900.0 - sy


def expand_path(d: str) -> list[tuple[str, list[float]]]:
    """Parse Klee SVG path d into a normalized command list using only M/L/C/Z.

    Expands H/V to L and converts Q to C. Returns a list of (op, args)
    where args is the flat list of numbers expected by the op.
    """
    tokens = re.findall(r"[MLCQHVZmlcqhvz]|-?\d+(?:\.\d+)?", d)
    out: list[tuple[str, list[float]]] = []
    i = 0
    cx = cy = 0.0
    while i < len(tokens):
        op = tokens[i]
        i += 1
        if op in "Zz":
            out.append(("Z", []))
            continue
        if op == "M":
            x = float(tokens[i]); i += 1
            y = float(tokens[i]); i += 1
            out.append(("M", [x, y]))
            cx, cy = x, y
        elif op == "L":
            x = float(tokens[i]); i += 1
            y = float(tokens[i]); i += 1
            out.append(("L", [x, y]))
            cx, cy = x, y
        elif op == "H":
            x = float(tokens[i]); i += 1
            out.append(("L", [x, cy]))
            cx = x
        elif op == "V":
            y = float(tokens[i]); i += 1
            out.append(("L", [cx, y]))
            cy = y
        elif op == "Q":
            p1x = float(tokens[i]); i += 1
            p1y = float(tokens[i]); i += 1
            p2x = float(tokens[i]); i += 1
            p2y = float(tokens[i]); i += 1
            c1x = cx + 2.0 / 3.0 * (p1x - cx)
            c1y = cy + 2.0 / 3.0 * (p1y - cy)
            c2x = p2x + 2.0 / 3.0 * (p1x - p2x)
            c2y = p2y + 2.0 / 3.0 * (p1y - p2y)
            out.append(("C", [c1x, c1y, c2x, c2y, p2x, p2y]))
            cx, cy = p2x, p2y
        elif op == "C":
            args = [float(tokens[i + k]) for k in range(6)]
            i += 6
            out.append(("C", args))
            cx, cy = args[4], args[5]
        else:
            raise ValueError(f"unsupported op {op!r} in path: {d!r}")
    return out


def render_d(cmds: list[tuple[str, list[float]]], fn, sep: str) -> str:
    """Serialize commands back into a path d, applying fn to each (x,y) pair."""
    parts: list[str] = []
    for op, args in cmds:
        parts.append(op)
        if op == "Z":
            continue
        pts = [(args[k], args[k + 1]) for k in range(0, len(args), 2)]
        for j, (x, y) in enumerate(pts):
            tx, ty = fn(x, y)
            if j > 0:
                parts.append(" ")
            parts.append(f"{round(tx)}{sep}{round(ty)}")
    return "".join(parts)


def render_median_d(pts: list[tuple[float, float]], fn) -> str:
    out: list[str] = []
    for j, (x, y) in enumerate(pts):
        tx, ty = fn(x, y)
        out.append(f"{'M' if j == 0 else 'L'}{round(tx)} {round(ty)}")
    return "".join(out)


def render_median_points(pts: list[tuple[float, float]], fn) -> list[list[int]]:
    return [[round(fn(x, y)[0]), round(fn(x, y)[1])] for x, y in pts]


def build_svg(codepoint: int, brush_d_svg: str, median_d_svg: str) -> str:
    cp = codepoint
    return (
        SVG_HEADER
        + f"\n<svg id=\"z{cp}\" class=\"acjk\" viewBox=\"0 0 1024 1024\" xmlns=\"http://www.w3.org/2000/svg\">\n"
        + SVG_STYLE
        + "\n"
        + f"<path id=\"z{cp}d1\" d=\"{brush_d_svg}\"/>\n"
        + "<defs>\n"
        + f"\t<clipPath id=\"z{cp}c1\"><use href=\"#z{cp}d1\"/></clipPath>\n"
        + "</defs>\n"
        + f"<path style=\"--d:1s;\" pathLength=\"3333\" clip-path=\"url(#z{cp}c1)\" d=\"{median_d_svg}\"/>\n"
        + "</svg>\n"
    )


def update_graphics_ja(entries: list[dict]) -> None:
    existing: list[str] = []
    if os.path.exists(GRAPHICS_FILE):
        with open(GRAPHICS_FILE, "r", encoding="utf-8") as f:
            existing = [line.rstrip("\n") for line in f if line.strip()]

    by_char = {json.loads(line)["character"]: line for line in existing}
    for entry in entries:
        by_char[entry["character"]] = json.dumps(entry, ensure_ascii=False, separators=(",", ":"))

    with open(GRAPHICS_FILE, "w", encoding="utf-8") as f:
        for line in by_char.values():
            f.write(line + "\n")


def main() -> int:
    font = TTFont(FONT_PATH)
    cmap = font["cmap"].getBestCmap()
    glyph_set = font.getGlyphSet()

    os.makedirs(SVGS_DIR, exist_ok=True)
    entries: list[dict] = []

    for cp, median_pts in MEDIANS.items():
        char = chr(cp)
        if cp not in cmap:
            print(f"skip U+{cp:04X} {char}: not in Klee", file=sys.stderr)
            continue
        name = cmap[cp]
        pen = SVGPathPen(glyph_set)
        glyph_set[name].draw(pen)
        d_font = pen.getCommands()
        cmds = expand_path(d_font)

        brush_svg = render_d(cmds, to_svg_xy, sep=" ")
        brush_gja = render_d(cmds, to_graphics_xy, sep=",")
        median_svg = render_median_d(median_pts, to_svg_xy)
        median_gja_pts = render_median_points(median_pts, to_graphics_xy)

        svg = build_svg(cp, brush_svg, median_svg)
        out_path = os.path.join(SVGS_DIR, f"{cp}.svg")
        with open(out_path, "w", encoding="utf-8") as f:
            f.write(svg)
        print(f"wrote {out_path}")

        entries.append({
            "character": char,
            "strokes": [brush_gja],
            "medians": [median_gja_pts],
        })

    update_graphics_ja(entries)
    print(f"updated {GRAPHICS_FILE} with {len(entries)} entries")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
