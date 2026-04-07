package ui

import "charm.land/lipgloss/v2"

var editorStyle = lipgloss.NewStyle().Padding(0, 1)

// View はエディタの描画内容を返す。
func (e *Editor) View() string {
	if e.noteID == "" {
		placeholder := "Press 'n' to create a note"
		if e.readOnly {
			placeholder = ""
		}

		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Width(e.width).
			Height(e.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(placeholder)
	}

	raw := e.textarea.View()
	if e.HasSelection() {
		raw = e.applySelectionHighlight(raw)
	}

	return editorStyle.Width(e.width).Height(e.height).Render(raw)
}
