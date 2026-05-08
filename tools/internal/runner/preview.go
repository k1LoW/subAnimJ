package runner

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
)

const previewTmpl = `<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>subAnimJ demo</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: sans-serif;
  padding: 2rem;
  max-width: 960px;
  margin: 0 auto;
  background: #fafafa;
  color: #222;
}
h1 { margin-bottom: 0.5rem; }
p.desc {
  color: #666;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
  line-height: 1.6;
}
p.desc a { color: #0366d6; }
.controls { margin-bottom: 1rem; }
button {
  padding: 6px 16px;
  font-size: 0.9rem;
  cursor: pointer;
  border: 1px solid #ccc;
  border-radius: 4px;
  background: #fff;
}
button:hover { background: #f0f0f0; }
.grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.cell { text-align: center; }
.cell .label {
  font-size: 1.2rem;
  font-weight: bold;
  margin-bottom: 4px;
}
.cell .meta {
  font-size: 0.75rem;
  color: #888;
  margin-bottom: 4px;
  font-family: monospace;
}
.cell .svg-container {
  width: 200px;
  height: 200px;
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
}
.cell .svg-container object {
  width: 100%;
  height: 100%;
  display: block;
}
</style>
</head>
<body>
<h1>subAnimJ demo</h1>
<p class="desc">
  Animated stroke-order SVGs of Japanese kanji modified for elementary
  school writing practice, derived from
  <a href="https://github.com/parsimonhi/animCJK">animCJK</a>
  (Arphic PL KaitiM via
  <a href="https://github.com/skishore/makemeahanzi">Make Me a Hanzi</a>).
  Inherits the
  <a href="https://ftp.gnu.org/non-gnu/chinese-fonts-truetype/LICENSE">Arphic Public License</a>.
  <a href="https://github.com/k1LoW/subAnimJ">GitHub</a>
</p>
<div class="controls">
  <button id="replay-btn">Replay</button>
</div>
<div class="grid" id="grid"></div>
<script>
const items = [
{{- range . }}
  { char: {{ .Char }}, codepoint: {{ .Codepoint }}, file: {{ .File }} },
{{- end }}
];
const grid = document.getElementById('grid');
function render() {
  grid.innerHTML = '';
  for (const it of items) {
    const cell = document.createElement('div');
    cell.className = 'cell';
    const label = document.createElement('div');
    label.className = 'label';
    label.textContent = it.char;
    cell.appendChild(label);
    const meta = document.createElement('div');
    meta.className = 'meta';
    meta.textContent = 'U+' + it.codepoint.toString(16).toUpperCase().padStart(4, '0');
    cell.appendChild(meta);
    const box = document.createElement('div');
    box.className = 'svg-container';
    const obj = document.createElement('object');
    obj.type = 'image/svg+xml';
    obj.data = it.file + '?t=' + Date.now();
    box.appendChild(obj);
    cell.appendChild(box);
    grid.appendChild(cell);
  }
}
document.getElementById('replay-btn').addEventListener('click', render);
render();
</script>
</body>
</html>
`

type previewItem struct {
	Char      string
	Codepoint int
	File      string
}

func writePreview(outRoot string, items []previewItem) error {
	tmpl, err := template.New("preview").Parse(previewTmpl)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, items); err != nil {
		return err
	}
	if err := os.MkdirAll(outRoot, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outRoot, "preview.html"), buf.Bytes(), 0o644)
}
