package runner

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
)

const previewTmpl = `<!doctype html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>subAnimJ Preview</title>
<style>
body {
  font-family: sans-serif;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}
h1 { text-align: center; }
.container {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 16px;
}
.kanji {
  text-align: center;
}
.kanji h2 {
  margin: 0 0 4px 0;
  font-size: 1.2em;
}
.kanji object {
  width: 200px;
  height: 200px;
  border: 1px solid #eee;
}
button {
  display: block;
  margin: 20px auto;
  padding: 8px 24px;
  font-size: 1em;
  cursor: pointer;
}
</style>
</head>
<body>
<h1>subAnimJ Preview</h1>
<button onclick="replay()">Replay</button>
<div class="container" id="container"></div>
<script>
const items = [
{{- range . }}
  { char: {{ .Char }}, file: {{ .File }} },
{{- end }}
];
const container = document.getElementById('container');
function loadAll() {
  container.innerHTML = '';
  items.forEach(it => {
    const div = document.createElement('div');
    div.className = 'kanji';
    const h2 = document.createElement('h2');
    h2.textContent = it.char;
    div.appendChild(h2);
    const obj = document.createElement('object');
    obj.type = 'image/svg+xml';
    obj.data = it.file + '?t=' + Date.now();
    div.appendChild(obj);
    container.appendChild(div);
  });
}
function replay() { loadAll(); }
loadAll();
</script>
</body>
</html>
`

type previewItem struct {
	Char string
	File string
}

func writePreview(outRoot string, pairs [][2]string) error {
	tmpl, err := template.New("preview").Parse(previewTmpl)
	if err != nil {
		return err
	}
	items := make([]previewItem, len(pairs))
	for i, p := range pairs {
		items[i] = previewItem{Char: p[0], File: p[1]}
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
