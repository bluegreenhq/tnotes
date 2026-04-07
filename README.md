# tnotes

ターミナルで使えるシンプルなメモ管理TUIアプリ。CLIモードにも対応。

![demo](demo.gif)

## インストール

```bash
brew install bluegreenhq/tap/tnotes
```

## 使い方

```bash
tnotes                  # TUIモードで起動
tnotes list             # ノート一覧を表示
tnotes get <id>         # 指定IDのノートを表示
tnotes help             # ヘルプを表示
```

操作方法の詳細は [docs/manual.md](docs/manual.md) を参照してください。

## コントリビューター向け

### セットアップ

```bash
make init   # 開発ツールのインストール + 依存取得
```

### 実行

```bash
make run                        # TUIモードで起動
make run ARGS=list              # ノート一覧を表示
make run ARGS="get <id>"        # 指定IDのノートを表示
make run ARGS=help              # ヘルプを表示
```

### 開発

```bash
make dev    # ホットリロード開発
make build  # ビルド
make test   # lint + 全テスト
make lint   # lint のみ
make demo   # デモGIF録画（要 vhs）
```
