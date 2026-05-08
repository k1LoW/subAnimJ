# subAnimJ

[animCJK](https://github.com/parsimonhi/animCJK) の日本の漢字データ (`svgsJa/` と `graphicsJa.txt`) を、日本の小学校の書き取り練習向けに改変したサブセットです。

## なぜ作るのか

animCJK の漢字 SVG はフォント由来のため、日本の小学校で教える字形と一部異なります。具体的には:

1. 「日」の 3 画目や「田」の 4 画目など、デザイン上は他のストロークと接続しない線が、日本の書き取り練習では他線まで届く字形が望まれる。
2. 「組」の糸偏のように中国字形・中国書き順となっており、日本の書き順 (赤い結び目を最後に書く順序) と一致しないものがある。

本リポジトリは、上流 animCJK のディレクトリ構成を踏襲しつつ、改変対象漢字の SVG と `graphicsJa.txt` の該当エントリだけを置きます。改変は決定的なツールで再現できるようにし、上流追従を容易にします。

## ディレクトリ構成

```
subAnimJ/
├── vendor/animCJK/        # 上流 (git submodule, コミットには含めない)
├── svgsJa/                # 改変済 SVG (該当字のみ, 上流と同じファイル名)
├── graphicsJa.txt         # 改変済エントリのみの JSONL (上流のサブセット)
├── parts/                 # 部品 SVG (糸偏など, 後続フェーズ)
├── tools/                 # 改変を行う Go ツール
├── targets/               # 改変指示 (1 漢字 1 ファイル)
├── preview.html           # 改変済 SVG をブラウザで一覧確認するためのページ
├── Makefile
├── CREDITS                # 上流クレジット・改変点
├── LICENSE                # svgsJa/ と graphicsJa.txt を覆う APL
└── tools/LICENSE          # 自前コード (tools/) の MIT
```

## 使い方

### 初回セットアップ

```sh
git clone https://github.com/k1LoW/subAnimJ.git
cd subAnimJ
git submodule update --init
```

### 改変済ファイルの再生成

```sh
make build
```

`targets/` 配下の `*.jsonl` を全て読み, 上流データから改変済 `svgsJa/*.svg`, `graphicsJa.txt`, および対応漢字を一覧する `preview.html` を再生成します。`preview.html` を任意の HTTP サーバーで開くと書き順アニメを確認できます (例: `python3 -m http.server` で開いて `http://localhost:8000/preview.html`)。

### 上流追従

```sh
make update-vendor
make build
```

## 改変方式

改変は `targets/{漢字}.jsonl` に **1 漢字 1 ファイル** で宣言します。1 ファイル内には複数の操作を 1 行 1 操作で記述でき, ファイル内の出現順に逐次適用されます。

`targets/日.jsonl` の例:

```jsonl
{"op":"extend","stroke":3,"direction":"horizontal"}
```

スキーマ:

| フィールド | 必須 | 内容 |
|---|---|---|
| `op` | yes | 現在は `extend` のみ |
| `stroke` | extend 時必須 | 1-origin のストローク番号 |
| `direction` | extend 時必須 | `horizontal` / `vertical` (画の主成分) |

### 延長 (`extend`) の挙動

1. 対象ストロークの中心線 (median) を取得し、両端のうち他ストロークから遠い方を「自由端」と判定。
2. 自由端の接線ベクトルを算出 (画の角度を保持)。
3. 接線方向に伸ばし、他ストロークの中心線と最初に交差する点で停止。
4. 中心線の自由端と、筆形ポリゴンの自由端寄り頂点を、同一の並進ベクトルで移動。

`direction` は接線が指定軸方向に支配的であることのサニティチェックに使われます (例: `horizontal` でほぼ垂直な画を指定するとエラー)。

## ツール開発

```sh
cd tools
go test ./...
go run ./extend --kanji 日 --stroke 3 --direction horizontal
```

## ライセンス・帰属

このリポジトリの主たる成果物である `svgsJa/*.svg` と `graphicsJa.txt` は、Arphic PL KaitiM フォントから派生した animCJK を更に派生させたもので、**Arphic Public License (APL)** が継承されます。リポジトリ直下の [`LICENSE`](./LICENSE) が APL 全文を含み、これらのファイルに適用されます。再配布時は APL 本文の同梱と Arphic Technology Co., Ltd. への帰属表示が必要です。

一方 `tools/` 以下の自前コードは MIT License です。コードのみを取り出して使う場合は [`tools/LICENSE`](./tools/LICENSE) に従ってください。

上流追跡や本リポジトリ独自の改変内容は [`CREDITS`](./CREDITS) にまとめています。上流 animCJK の作者 FM&SH 氏に感謝します。
