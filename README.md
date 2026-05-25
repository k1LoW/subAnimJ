# subAnimJ

Modifications to a small selection of files in
[animCJK](https://github.com/parsimonhi/animCJK)'s Japanese kanji data
(`svgsJa/` and `graphicsJa.txt`), tailored for Japanese elementary school
writing practice.

## Demo

- Modified kanji: <https://k1low.github.io/subAnimJ/>

## Why

animCJK's kanji SVGs are font-derived, so a few glyph shapes diverge from how
characters are taught in Japanese elementary schools:

1. Strokes that, by font design, do not visually meet adjacent strokes (the
   middle horizontal of `日` / `田`, etc.) are expected to reach those strokes
   in handwriting practice.
2. Some radicals — notably the thread radical (糸偏) and the bamboo radical
   (竹冠) — use a Chinese-style form with stroke shape and stroke order that
   differ from what is taught in Japan.

This repository keeps the upstream directory layout but only carries the
modified SVG files and the corresponding entries in `graphicsJa.txt`.
Modifications are applied by a deterministic Go tool so the output can be
regenerated whenever upstream changes.

## Layout

```
subAnimJ/
├── vendor/animCJK/        # upstream (git submodule, not committed directly)
├── svgsJa/                # modified SVG files (one per affected kanji)
├── graphicsJa.txt         # modified entries only (subset of upstream JSONL)
├── parts/                 # hand-authored Japanese-style radical SVGs (糸, 竹, ...)
├── tools/                 # Go tool that performs the modifications
├── targets/               # modification specs, one JSONL file per kanji
├── preview.html           # locally generated browser preview
├── Makefile
├── CREDITS                # upstream attribution and modification notes
├── LICENSE                # APL, covers svgsJa/ and graphicsJa.txt
└── tools/LICENSE          # MIT, covers the Go code under tools/
```

## Usage

### First-time setup

```sh
git clone https://github.com/k1LoW/subAnimJ.git
cd subAnimJ
git submodule update --init
```

### Regenerate the modified files

```sh
make build
```

This reads every `*.jsonl` under `targets/`, applies the operations against
the upstream data, and writes `svgsJa/*.svg`, the `graphicsJa.txt` subset,
and `preview.html`. Open `preview.html` through any HTTP server to see the
stroke-order animations (for example `python3 -m http.server` and visit
`http://localhost:8000/preview.html`).

### Track upstream

```sh
make update-vendor
make build
```

## Specifying modifications

Modifications are declared in `targets/{kanji}.jsonl`, **one JSONL file per
kanji**. A file may contain multiple operations, one per line; operations
within a file are applied in order so each step sees the cumulative effect of
the previous ones.

Example `targets/日.jsonl`:

```jsonl
{"op":"extend","stroke":3,"direction":"horizontal"}
```

Example `targets/組.jsonl` (compose then extend):

```jsonl
{"op":"compose","part":"糸","fit_stroke":1}
{"op":"extend","stroke":9,"direction":"horizontal"}
{"op":"extend","stroke":10,"direction":"horizontal"}
```

Schema:

| field | applies to | meaning |
|---|---|---|
| `op` | all | `extend` or `compose` |
| `stroke` | `extend` | 1-origin stroke index in upstream order |
| `direction` | `extend` | `horizontal` or `vertical` (sanity check; at least one end's tangent must match) |
| `side` | `extend` | `start`, `end`, or `both` (default `both`); explicit sides bypass the brush-gap threshold |
| `part` | `compose` | name under `parts/` (without the `.svg` extension) |
| `start` | `compose` | first upstream stroke to replace (1-origin, default 1) |
| `fit_stroke` | `compose` | 1-origin part-stroke index whose bbox is matched against the corresponding upstream stroke; the resulting affine transform applies to every part stroke |
| `offset` | `compose` | `[dx, dy]` applied after the fit transform for fine-tuning placement; `-x` is left, `+y` is up (graphicsJa y-up). Example: `"offset":[-30,0]` shifts the composed radical 30 units left |

### How `extend` works

For each end (start and end) of the target stroke:

1. Compute the tangent at that end from the median centerline so the original
   stroke angle is preserved.
2. Cast a ray from the end along the tangent; find the nearest intersection
   with any other stroke's median.
3. If the gap to that intersection is more than ~30 units, translate the end
   by `tangent * gap`. Brush polygon vertices on the same half (closer to that
   end than to the opposite end) move with it. When `side` is `start` or
   `end`, the threshold is bypassed so a marginally short end still extends.

The 30-unit threshold treats already-overlapping strokes (such as the left
end of the middle horizontal of `日`, which sits inside the left vertical's
brush stroke) as already connected and leaves them in place. Strokes with a
real visible gap (such as both ends of the middle horizontal of `田`) get
extended on both sides automatically.

### How `compose` works

`compose` replaces a contiguous run of strokes — typically a radical — with
a hand-authored part SVG from `parts/`. Each part SVG follows the same
structure as a normal animCJK glyph (one `<path id="dN">` per stroke plus
the clipped median path) but is drawn in Japanese style.

1. Load `parts/{part}.svg` and read its N strokes (brush polygon plus
   median polyline).
2. Compute the bounding box of the part's `fit_stroke` (default the whole
   part if zero/omitted) and the corresponding upstream stroke at
   `start + fit_stroke - 1`.
3. Derive an affine transform (independent x/y scale plus translate) that
   maps the source bbox onto the target bbox.
4. Apply the transform to every stroke in the part and overwrite the
   upstream's strokes `[start, start+N-1]` with the result. If `offset`
   is given, every transformed point is also translated by `[dx, dy]`
   for fine-tuning placement. The total stroke count is preserved, so
   later strokes' indices (e.g. the 且 part in `組`) keep their original
   numbering.

Using `fit_stroke=1` is the recommended default for radicals like 糸 and 竹
where the first stroke is positionally consistent between the Chinese and
Japanese forms; this keeps the radical's internal proportions constant
while still anchoring it to the upstream's layout. Use `offset` when the
fit-derived position needs a small nudge (for example, `筆`'s 竹冠 sits
a little far right when fitted naturally, so `"offset":[-30,0]` brings it
back).

## Tool development

```sh
cd tools
go test ./...
go run ./extend --kanji 日 --stroke 3 --direction horizontal
```

## License and attribution

The kanji SVG files under `svgsJa/` (everything except the eight Japanese
punctuation glyphs listed below) and the corresponding entries in
`graphicsJa.txt` are derived from animCJK, which itself derives from the
Arphic PL KaitiM fonts. They inherit the **Arphic Public License (APL)**.
The repository-root [`LICENSE`](./LICENSE) contains the full APL text. When
redistributing, include the APL text and attribute Arphic Technology Co.,
Ltd.

The eight Japanese punctuation SVGs (`12289.svg`, `12290.svg`, `12300.svg`,
`12301.svg`, `65041.svg`, `65042.svg`, `65089.svg`, `65090.svg`) and their
`graphicsJa.txt` entries are generated by `tools/extract_punct.py` from the
[Klee One](https://github.com/fontworks-fonts/Klee) font. They inherit the
**SIL Open Font License, Version 1.1**. See [`licenses/OFL.txt`](./licenses/OFL.txt)
for the full text and attribute The Klee Project Authors (Fontworks Inc.)
when redistributing.

The original Go code under `tools/` is licensed under the MIT License. If
you extract only the code, follow [`tools/LICENSE`](./tools/LICENSE).

[`CREDITS`](./CREDITS) lists upstream sources, license details, and the
modifications applied here. Thanks to the animCJK author FM&SH, the
Arphic / Make Me a Hanzi communities, and the Klee One authors whose work
this builds on.
