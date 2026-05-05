# CLAUDE.md

## パッケージ依存ルール

`internal/` 配下のパッケージ依存は以下の方向のみ許可する。depguard でも import レベルで強制している。

```
ui  → app, note, ui/shared
ui/shared → （internal の他パッケージに依存しない。dogubako/tui や標準ライブラリのみ）
app → store, note
store → note
```

- 上記以外の依存は禁止（import だけでなく、公開APIの引数・返り値の型としても露出させない）
- 許可されていないパッケージの機能が必要な場合は、依存可能な中間層にメソッドを追加して経由する
- `internal/ui/shared` は dogubako/tui への昇格候補のステージング層。internal の他パッケージに依存できない（depguard で強制）

## UIコンポーネントのファイル分割ルール

`internal/ui/` 配下のコンポーネントは原則として1ファイル（`xxx.go`）に統一する。

## 主要構造体の責務

### note パッケージ

| 構造体 | 責務 |
|---|---|
| `Note` | ノート1件のドメインモデル。Metadata（ID・タイトル・パス等）と Body を保持。フォルダ判定・タイトル導出などの純粋なビジネスロジックを持つ |

### store パッケージ

| 構造体 | 責務 |
|---|---|
| `Store` | 永続化のインターフェース。CRUD・ゴミ箱・フォルダ操作・import/export を定義 |
| `FileStore` | Store のファイルシステム実装。`index.json`（メタデータ一覧）とノートファイル（frontmatter + body）でデータを管理 |

### app パッケージ

| 構造体 | 責務 |
|---|---|
| `App` | アプリケーションロジック層。インメモリのノート一覧を保持し、Store を介して永続化する。ノートのCRUD・フォルダ操作・検索・undo/redo を提供。UI と CLI の両方から利用される |
| `NoteUndoManager` | ノート操作（作成・削除・複製）の undo/redo スタック管理 |
| `NoteResult` | ノート操作の結果を UI に伝達するための DTO。操作後のノート一覧・選択インデックス・ヒントメッセージを含む |

### ui パッケージ

| 構造体 | 責務 |
|---|---|
| `Model` | bubbletea のトップレベル Model。全 UI コンポーネントを保持し、Update/View を統括する |
| `Layout` | 画面レイアウトの座標計算。各領域の幅・開始位置・ヒットテスト（クリック座標→領域判定）を担う |
| `FolderList` | フォルダ一覧ペイン。Notes/ユーザー定義フォルダ/Trash の表示・選択・インライン入力（新規作成・リネーム）を管理 |
| `NoteList` | ノート一覧ペイン。セクション分け（Today/Yesterday 等）・スクロール・選択を管理 |
| `Editor` | テキスト編集ペイン。simpleTextArea のラッパーで、選択範囲・undo/redo・右クリックメニュー・検索ハイライトを統合 |
| `EditorHeader` | エディタ上部のヘッダー。新規作成/…メニューボタン・検索フィールド・移動先フォルダメニューを管理 |
| `simpleTextArea` | 独自テキストエリア。行バッファ（`[][]rune`）・カーソル位置・スクロールオフセットを保持。ソフトラップ/ノーラップのレイアウト戦略を持つ |
| `EditorUndoManager` | エディタ内テキスト編集の undo/redo スタック管理。スナップショットを一定間隔で記録 |
| `Footer` | フッターバー。メニューボタンと PopupMenu を管理 |
| `PopupMenu` | 汎用ポップアップメニュー。項目リスト・ホバー状態・クリック判定を持つ再利用コンポーネント |
| `PopupCoordinator` | 複数コンポーネントに散在する PopupMenu を一元管理。排他制御・アンカー位置計算・キー/クリック入力のディスパッチを担う |
| `ConfirmDialog` | 汎用確認ダイアログ（Yes/No）。削除確認等に使用 |
| `HelpOverlay` | ショートカットキー一覧オーバーレイ。フォーカス領域に応じた内容を表示 |

### cli パッケージ

| 構造体 | 責務 |
|---|---|
| （関数ベース） | CLI サブコマンドのディスパッチと実行。`app.App` を介してノート操作を行い、結果を stdout に出力する |

## ルート Model の設計方針

bubbletea のルート `Model` は可能な限りシンプルに保つ。肥大化を防ぐために以下を適宜検討する。

- 子コンポーネントの切り出し
- ビジネスロジックの app layer への抽出
- ユーティリティ関数への抽出
- 子コンポーネントへの処理委譲

## View の純粋性

`View()` メソッドは描画専用とし、状態変更を行わない。状態変更は `Update` （または `Init`）経路でのみ行う。

- `View()` 内でのフィールド書き込み・セッター呼び出しは禁止
- 描画に必要な事前計算（オフセット調整・サイズ設定等）は `Update` 側で行う
- 判定関数（`NeedsXxx` 等）に副作用を持たせない

## Update 内の I/O

- ネットワーク通信など遅い I/O は `tea.Cmd` で非同期に行う
- ローカルファイル書き込み（`Save` 等）は `Update` 内で同期的に行ってよい

## WithStack の使い分け

`errors.WithStack` は標準ライブラリなどスタックトレースを持たないエラーにのみ使用する。内部関数や `cockroachdb/errors` で生成済みのエラーには不要。

## 公開・非公開の方針

internal パッケージ内のメソッド・関数は、自身の receiver からしか呼ばれないものは private（小文字始まり）にする。他の型や関数から呼ばれるもの、テスト（`package ui_test`）から呼ばれるもの、インターフェース実装は public のままにする。

## デモ録画の制約

`demo.exp`（expect スクリプト）で asciinema 録画を行う際、Shift+Arrow キー（`\x1b[1;2B` 等）による範囲選択は bubbletea の Kitty キーボードプロトコル有効化環境で機能しない。デモでは Shift+Arrow の代わりに `Ctrl+Shift+A`（全選択）や `Ctrl+K`/`Ctrl+Y`（kill/yank）を使うこと。