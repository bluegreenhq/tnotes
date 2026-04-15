package screencopy

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/x/ansi"
)

const flashDuration = 2 * time.Second

// clearFlashMsg はフラッシュメッセージを消すための内部メッセージ。
type clearFlashMsg struct {
	id int
}

// Model は任意の tea.Model をラップし、Ctrl+Shift+4 で画面コピー機能を追加する。
type Model struct {
	inner        tea.Model
	lastRendered string
	flashMsg     string
	flashID      int
}

// Wrap は inner を画面コピー機能付きでラップする。
func Wrap(inner tea.Model) *Model {
	return &Model{inner: inner, lastRendered: "", flashMsg: "", flashID: 0}
}

// Init は inner の Init に委譲する。
func (m *Model) Init() tea.Cmd {
	return m.inner.Init()
}

// Update はキーイベントを処理し、Ctrl+Shift+4 でクリップボードにコピーする。
// それ以外のメッセージは inner に委譲する。
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.Code == '4' && msg.Mod == (tea.ModCtrl|tea.ModShift) {
			return m, m.copyScreen()
		}
	case clearFlashMsg:
		if msg.id == m.flashID {
			m.flashMsg = ""
		}

		return m, nil
	}

	inner, cmd := m.inner.Update(msg)
	m.inner = inner

	return m, cmd
}

// View は inner の View を呼び出してキャッシュし、フラッシュメッセージがあればオーバーレイする。
func (m *Model) View() tea.View {
	v := m.inner.View()
	m.lastRendered = v.Content

	if m.flashMsg != "" {
		v.Content = m.overlayFlash(v.Content)
	}

	return v
}

func (m *Model) copyScreen() tea.Cmd {
	if m.lastRendered == "" {
		return nil
	}

	err := clipboard.WriteAll(m.lastRendered)
	if err != nil {
		return m.setFlash("Failed to copy screen")
	}

	return m.setFlash("Screen copied to clipboard")
}

func (m *Model) setFlash(msg string) tea.Cmd {
	m.flashID++
	m.flashMsg = msg

	id := m.flashID

	return tea.Tick(flashDuration, func(_ time.Time) tea.Msg {
		return clearFlashMsg{id: id}
	})
}

func newFlashStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 1)
}

// overlayFlash は画面右上にフラッシュメッセージをオーバーレイする。
func (m *Model) overlayFlash(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	rendered := newFlashStyle().Render(m.flashMsg)
	msgWidth := lipgloss.Width(rendered)

	firstLineWidth := lipgloss.Width(lines[0])
	startX := max(firstLineWidth-msgWidth, 0)

	truncated := ansi.Truncate(lines[0], startX, "")
	w := lipgloss.Width(truncated)
	padded := truncated + strings.Repeat(" ", startX-w)
	lines[0] = padded + rendered

	return strings.Join(lines, "\n")
}
