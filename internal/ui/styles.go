package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
)

var (
	buttonStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	buttonHoverStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))

	editorStyle = lipgloss.NewStyle().Padding(0, 1)

	editorBoldStyle      = lipgloss.NewStyle().Bold(true)
	editorCursorStyle    = lipgloss.NewStyle().Reverse(true)
	editorSelectionStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15"))

		// タイトル太字.
	editorBoldOn  = shared.ExtractANSIOn(editorBoldStyle)
	editorBoldOff = shared.ExtractANSIOff(editorBoldStyle)

	// カーソル: reverse のみをトグルし、他の属性を維持する。
	editorCursorOn  = shared.ExtractANSIOn(editorCursorStyle)
	editorCursorOff = shared.ExtractANSIOff(editorCursorStyle)

	// 選択: fg/bg を設定し、解除時は fg/bg のみリセットする。
	editorSelectionOn  = shared.ExtractANSIOn(editorSelectionStyle)
	editorSelectionOff = shared.ExtractANSIOff(editorSelectionStyle)

	// 検索ハイライト: ノート一覧（背景黄色 + 黒文字、エディタと統一）.
	searchHighlightNoteListStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("3")).
					Foreground(lipgloss.Color("0")).
					Bold(true)
	searchHighlightNoteListSelectedStyle = lipgloss.NewStyle().
						Background(lipgloss.Color("3")).
						Foreground(lipgloss.Color("0")).
						Bold(true)

	// 検索ハイライト: エディタ（背景黄色 + 黒文字）.
	editorSearchHighlightStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("3")).
					Foreground(lipgloss.Color("0"))

	editorSearchHighlightOn  = shared.ExtractANSIOn(editorSearchHighlightStyle)
	editorSearchHighlightOff = shared.ExtractANSIOff(editorSearchHighlightStyle)
)
