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
2. Some radicals, notably the thread radical (糸偏), use a Chinese-style form
   with a stroke order that differs from what is taught in Japan.

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
├── parts/                 # custom radical SVGs (future phase)
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

Schema:

| field | required | meaning |
|---|---|---|
| `op` | yes | currently only `extend` |
| `stroke` | for `extend` | 1-origin stroke index (matches upstream order) |
| `direction` | for `extend` | `horizontal` or `vertical` (the stroke's main axis) |

### How `extend` works

For each end (start and end) of the target stroke:

1. Compute the tangent at that end from the median centerline so the original
   stroke angle is preserved.
2. Skip the end if its tangent does not match the requested `direction`
   (example: a `horizontal` request will not extend a vertical end).
3. Cast a ray from the end along the tangent; find the nearest intersection
   with any other stroke's median.
4. If the gap to that intersection is more than ~30 units, translate the end
   by `tangent * gap`. Brush polygon vertices on the same half (closer to that
   end than to the opposite end) move with it.

The 30-unit threshold treats already-overlapping strokes (such as the left
end of the middle horizontal of `日`, which sits inside the left vertical's
brush stroke) as already connected and leaves them in place. Strokes with a
real visible gap (such as both ends of the middle horizontal of `田`) get
extended on both sides.

## Tool development

```sh
cd tools
go test ./...
go run ./extend --kanji 日 --stroke 3 --direction horizontal
```

## License and attribution

The primary deliverables of this repository, `svgsJa/*.svg` and
`graphicsJa.txt`, are derived from animCJK, which itself derives from the
Arphic PL KaitiM fonts. They inherit the **Arphic Public License (APL)**.
The repository-root [`LICENSE`](./LICENSE) contains the full APL text and
applies to those files. When redistributing, include the APL text and
attribute Arphic Technology Co., Ltd.

The original Go code under `tools/` is licensed under the MIT License. If you
extract only the code, follow [`tools/LICENSE`](./tools/LICENSE).

[`CREDITS`](./CREDITS) lists upstream sources, license details, and the
modifications applied here. Thanks to the animCJK author FM&SH and the
Arphic / Make Me a Hanzi communities whose work this builds on.
