# Notes

ターミナルで使えるシンプルなメモ管理TUIアプリ。CLIモードにも対応。

![demo](demo.gif)

## 使い方

```bash
make run                        # TUIモードで起動
make run ARGS=list              # ノート一覧を表示
make run ARGS="get <id>"        # 指定IDのノートを表示
make run ARGS=help              # ヘルプを表示
```

操作方法の詳細は [docs/manual.md](docs/manual.md) を参照してください。

## 開発

```bash
make test   # lint + 全テスト
make lint   # lint のみ
make build  # ビルド
make dev    # ホットリロード開発
make demo   # デモGIF録画（要 vhs）
```
