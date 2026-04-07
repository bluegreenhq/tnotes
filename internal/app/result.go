package app

import "github.com/bluegreenhq/tnotes/internal/note"

// NoteResult はノート操作の結果をUIに伝達するための構造体。
type NoteResult struct {
	// Note は操作対象のノート。作成・復元時にエディタへ読み込むために使用する。
	// ゴミ箱移動等、対象ノートの内容が不要な場合はゼロ値。
	Note note.Note
	// Notes は操作後のメインノート一覧。
	Notes []note.Note
	// SelectIdx はUI側が選択すべきノートのインデックス。-1 はノートなし。
	SelectIdx int
	// InfoHint はフッターに表示するヒントメッセージ。空文字列は表示しない。
	InfoHint string
}
