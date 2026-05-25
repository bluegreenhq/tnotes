package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
)

// ConfirmDialogOverlay は tui.ConfirmDialog をオーバーレイ化したラッパー。
// Yes 選択時に onConfirm を実行する。
type ConfirmDialogOverlay struct {
	dialog    *tui.ConfirmDialog
	onConfirm ModelAction
}

var _ shared.OverlayComponent = (*ConfirmDialogOverlay)(nil)

// NewConfirmDeleteFolderDialog はフォルダ削除確認用の ConfirmDialogOverlay を生成する。
func NewConfirmDeleteFolderDialog(name string, noteCount int) *ConfirmDialogOverlay {
	title := fmt.Sprintf("Delete %q?", name)
	detail := fmt.Sprintf("%d note(s) will be moved to Trash.", noteCount)
	d := tui.NewConfirmDialog(title, detail)

	return &ConfirmDialogOverlay{dialog: &d, onConfirm: confirmDeleteFolder(name)}
}

// SetScreenSize は画面サイズを設定する。
func (c *ConfirmDialogOverlay) SetScreenSize(width, height int) {
	c.dialog.SetScreenSize(width, height)
}

// UpdateOverlay はメッセージに応じて状態を更新し、ModelAction と tea.Cmd を返す。
func (c *ConfirmDialogOverlay) UpdateOverlay(msg tea.Msg) (ModelAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return c.resultAction(c.dialog.Update(msg)), nil
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return c.resultAction(c.dialog.HandleClickAbs(msg.X, msg.Y)), nil
		}
	case tea.MouseMsg:
		mouse := msg.Mouse()
		c.dialog.HandleMotionAbs(mouse.X, mouse.Y)
	}

	return nil, nil
}

// RenderOn はベース画面上に確認ダイアログを合成する。
func (c *ConfirmDialogOverlay) RenderOn(base string, width, height int) string {
	bodyLines := strings.Split(base, "\n")
	rendered := c.dialog.View()
	g := tui.CalcOverlayGeometry(rendered, width, height, 0, 0, 0)
	tui.OverlayLines(bodyLines, strings.Split(rendered, "\n"), g.StartX, g.StartY)

	return strings.Join(bodyLines, "\n")
}

func (c *ConfirmDialogOverlay) resultAction(result tui.ConfirmResult) ModelAction {
	switch result {
	case tui.ConfirmYes:
		return c.dismissAndRun(true)
	case tui.ConfirmNo:
		return c.dismissAndRun(false)
	case tui.ConfirmContinue:
		return nil
	}

	return nil
}

// dismissAndRun は overlay をクリアし、Yes ならば onConfirm を実行する。
func (c *ConfirmDialogOverlay) dismissAndRun(confirmed bool) ModelAction {
	return func(m *Model, ctx ActionContext) tea.Cmd {
		m.Overlays.Clear()

		if !confirmed || c.onConfirm == nil {
			return nil
		}

		return c.onConfirm(m, ctx)
	}
}

// --- ConfirmDialogOverlay → Model アクション ---

// confirmDeleteFolder はフォルダ削除確認の Yes 押下時に呼ばれる。
func confirmDeleteFolder(name string) ModelAction {
	return func(m *Model, ctx ActionContext) tea.Cmd {
		deleted, err := m.FolderList.DeleteFolder(name)
		if err != nil {
			return reportResult(err, "")(m, ctx)
		}

		info := "Deleted: " + name + " (" + strconv.Itoa(deleted) + " note(s) trashed)"

		return reportResult(nil, info)(m, ctx)
	}
}
