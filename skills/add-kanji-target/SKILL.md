---
name: add-kanji-target
description: >-
  Add or update a kanji modification spec under `targets/` and rebuild.
  Recognizes the user's terse input format:
    "漢字 N[,N2,...] 横|縦 [start|end]"  → extend op(s)
    "漢字 糸偏|糸編|竹冠|女偏|<part>"      → compose op (compose then optional extends)
  Combine forms like "漢字 9,10 横 + 糸偏" stack both compose and extend
  in the same JSONL file (compose first, then extends).
  Use when the user types a line matching that shape during the
  per-kanji modification workflow.
---

# Add kanji target

Workflow for the per-kanji modification loop. The user supplies a terse
spec; this skill produces or updates `targets/{kanji}.jsonl`, runs
`make build`, and reports the result. The kanji's stroke order follows
**upstream animCJK numbering** (which is Chinese order, sometimes
different from Japanese order — see step 3 below).

## Input parsing

Examples (all single-line):

| User input | Meaning |
|---|---|
| `日 3 横` | extend stroke 3 horizontally (both sides default) |
| `馬 4,5 横 start` | extend strokes 4 and 5 horizontally, start side only |
| `当 5 横 end` | extend stroke 5 horizontally, end side only |
| `歌 5 縦` | extend stroke 5 vertically |
| `絵 糸偏` | compose 糸 part over strokes 1-6 |
| `筆 竹冠` | compose 竹 part over strokes 1-6 |
| `姉 女偏` | compose 女 part over strokes 1-3 |
| `組 9,10 横 + 糸偏` | compose 糸 first, then extend strokes 9 and 10 |

Aliases:
- 横 / horizontal / h
- 縦 / vertical / v
- start / s = side "start"
- end / e = side "end"
- 糸編 = 糸偏 (typo-tolerant) → part "糸"
- 竹冠 → part "竹"
- 女偏 → part "女"
- generic `<part>` (e.g. `辶`, `阝`, …) → part "<part>" (assumes parts/<part>.svg exists)

## Steps

1. **Sanity check the kanji**:
   ```sh
   grep '"character":"<漢字>"' vendor/animCJK/graphicsJa.txt \
     | python3 -c "import sys, json; d=json.loads(sys.stdin.read()); [print(f'stroke {i+1}: {m}') for i, m in enumerate(d['medians'])]"
   ```
   If the kanji isn't found, surface the error.

2. **Validate each `extend` target**:
   - Compute the stroke's median tangent at start and end.
   - If `direction` is `horizontal`, at least one end must have `|tangent.x| > |tangent.y|`. If neither does (the stroke is purely vertical / a 折れ whose only horizontal end is opposite), report and propose the upstream index that matches.
   - Common gotcha: `田` / `町` / `画` / `角` etc. use the Chinese stroke order where stroke 3 is the middle vertical and stroke 4 (or later) is the middle horizontal. If the user-supplied "N 横" lands on a vertical, look for the nearest horizontal at N+1 or N-1 and confirm with the user before substituting.

3. **Write the targets file**:
   - Path: `targets/<漢字>.jsonl`
   - Order operations: `compose` first (if any), then `extend` in user-supplied order.
   - Default `fit_stroke` for compose is `1`.

   Example for `組 9,10 横 + 糸偏`:
   ```jsonl
   {"op":"compose","part":"糸","fit_stroke":1}
   {"op":"extend","stroke":9,"direction":"horizontal"}
   {"op":"extend","stroke":10,"direction":"horizontal"}
   ```

4. **Build and confirm**:
   ```sh
   make build 2>&1 | grep -E "<漢字>|Error"
   ```
   - Success line `built: <漢字>.jsonl` → report done.
   - Error like `no end could be extended in direction ...` → revisit step 2 (likely stroke-index/direction mismatch); ask the user to clarify rather than silently substituting.

5. **Brief reply**:
   - One line per kanji: which strokes were extended, and the new median endpoint(s) if helpful for verification.
   - Do NOT commit. The user runs `/commit` when they want to commit a batch.

## Notes

- `parts/<name>.svg` must exist for `compose`. If it doesn't, surface that instead of writing a broken target.
- When the user supplies multiple kanji in one prompt (e.g. `田 4 横` followed by `町 4 横`), process them sequentially.
- The HTTP preview server is typically already running on port 8773; don't restart it unless the user asks.
