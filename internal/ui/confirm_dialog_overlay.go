package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

// ConfirmTarget は確認ダイアログの対象を表す。
type ConfirmTarget int

const (
	// ConfirmTargetNone は対象未指定。
	ConfirmTargetNone ConfirmTarget = iota
	// ConfirmTargetFolderDelete はフォルダ削除確認。
	ConfirmTargetFolderDelete
)

// ConfirmDialogResultMsg は確認ダイアログの結果を通知する。
type ConfirmDialogResultMsg struct {
	Target    ConfirmTarget
	Confirmed bool
	// FolderName は ConfirmTargetFolderDelete の対象フォルダ名。
	FolderName string
}

// ConfirmDialogOverlay は tui.ConfirmDialog をオーバーレイ化したラッパー。
type ConfirmDialogOverlay struct {
	dialog     *tui.ConfirmDialog
	target     ConfirmTarget
	folderName string
}

var _ OverlayComponent = (*ConfirmDialogOverlay)(nil)

// NewConfirmDeleteFolderDialog はフォルダ削除確認用の ConfirmDialogOverlay を生成する。
func NewConfirmDeleteFolderDialog(name string, noteCount int) *ConfirmDialogOverlay {
	title := fmt.Sprintf("Delete %q?", name)
	detail := fmt.Sprintf("%d note(s) will be moved to Trash.", noteCount)
	d := tui.NewConfirmDialog(title, detail)

	return &ConfirmDialogOverlay{dialog: &d, target: ConfirmTargetFolderDelete, folderName: name}
}

// SetScreenSize は画面サイズを設定する。
func (c *ConfirmDialogOverlay) SetScreenSize(width, height int) {
	c.dialog.SetScreenSize(width, height)
}

// Update はメッセージに応じて状態を更新する。
func (c *ConfirmDialogOverlay) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return c.resultCmd(c.dialog.Update(msg))
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return c.resultCmd(c.dialog.HandleClickAbs(msg.X, msg.Y))
		}
	case tea.MouseMsg:
		mouse := msg.Mouse()
		c.dialog.HandleMotionAbs(mouse.X, mouse.Y)
	}

	return nil
}

// RenderOn はベース画面上に確認ダイアログを合成する。
func (c *ConfirmDialogOverlay) RenderOn(base string, width, height int) string {
	bodyLines := strings.Split(base, "\n")
	rendered := c.dialog.View()
	g := tui.CalcOverlayGeometry(rendered, width, height, 0, 0, 0)
	tui.OverlayLines(bodyLines, strings.Split(rendered, "\n"), g.StartX, g.StartY)

	return strings.Join(bodyLines, "\n")
}

// Target は確認対象を返す。
func (c *ConfirmDialogOverlay) Target() ConfirmTarget { return c.target }

// FolderName は対象フォルダ名を返す。
func (c *ConfirmDialogOverlay) FolderName() string { return c.folderName }

func (c *ConfirmDialogOverlay) resultCmd(result tui.ConfirmResult) tea.Cmd {
	switch result {
	case tui.ConfirmYes:
		target := c.target
		name := c.folderName

		return func() tea.Msg {
			return ConfirmDialogResultMsg{Target: target, Confirmed: true, FolderName: name}
		}
	case tui.ConfirmNo:
		target := c.target
		name := c.folderName

		return func() tea.Msg {
			return ConfirmDialogResultMsg{Target: target, Confirmed: false, FolderName: name}
		}
	case tui.ConfirmContinue:
		return nil
	}

	return nil
}
