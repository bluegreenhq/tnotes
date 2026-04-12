package ui

// HelpItem はショートカット1件を表す。
type HelpItem struct {
	Key         string
	Description string
}

// HelpSection はショートカットのグループを表す。
type HelpSection struct {
	Title string
	Items []HelpItem
}

// HelpOverlay はショートカット一覧オーバーレイコンポーネント。
type HelpOverlay struct {
	sections   []HelpSection
	closeHover bool
}

// NewHelpOverlay はフォーカスに応じた HelpOverlay を生成する。
func NewHelpOverlay(focus FocusArea) *HelpOverlay {
	var sections []HelpSection

	switch focus {
	case FocusNoteList:
		sections = append(sections, noteListHelpSection())
	case FocusFolderList:
		sections = append(sections, folderListHelpSection())
	case FocusEditor:
		sections = append(sections, editorHelpSection())
	}

	sections = append(sections, globalHelpSection())

	return &HelpOverlay{sections: sections, closeHover: false}
}

func noteListHelpSection() HelpSection {
	return HelpSection{
		Title: "Note List",
		Items: []HelpItem{
			{"j/↓/Ctrl+N", "Move down"},
			{"k/↑/Ctrl+P", "Move up"},
			{"n", "New note"},
			{"d/Del", "Delete note"},
			{"m", "Menu"},
			{"Enter", "Edit note"},
			{"Ctrl+C", "Copy note"},
			{"Ctrl+D", "Duplicate note"},
			{"Ctrl+Z", "Undo"},
			{"Ctrl+Shift+Z", "Redo"},
			{"q", "Quit"},
		},
	}
}

func folderListHelpSection() HelpSection {
	return HelpSection{
		Title: "Folder List",
		Items: []HelpItem{
			{"j/↓/Ctrl+N", "Move down"},
			{"k/↑/Ctrl+P", "Move up"},
			{"m", "Menu"},
			{"Enter/Tab", "Next pane"},
			{"q", "Quit"},
		},
	}
}

func editorHelpSection() HelpSection {
	return HelpSection{
		Title: "Editor",
		Items: []HelpItem{
			{"Ctrl+S", "Save"},
			{"Ctrl+Z", "Undo"},
			{"Ctrl+Shift+Z", "Redo"},
			{"Ctrl+Shift+A", "Select all"},
			{"Ctrl+C", "Copy"},
			{"Ctrl+X", "Cut"},
			{"Ctrl+V", "Paste"},
			{"Ctrl+O", "Open URL"},
			{"Ctrl+A", "Beginning of line"},
			{"Ctrl+E", "End of line"},
			{"Ctrl+F", "Move right"},
			{"Ctrl+B", "Move left"},
			{"Ctrl+N", "Move down"},
			{"Ctrl+P", "Move up"},
			{"Ctrl+D", "Delete character"},
			{"Ctrl+K", "Kill to end of line"},
			{"Ctrl+Y", "Yank"},
			{"Shift+Arrow", "Select"},
			{"Esc/Tab", "Back to list"},
		},
	}
}

func globalHelpSection() HelpSection {
	return HelpSection{
		Title: "Global",
		Items: []HelpItem{
			{"Ctrl+Q", "Quit"},
			{"Ctrl+B", "Toggle folders"},
			{"Tab", "Next pane"},
		},
	}
}
