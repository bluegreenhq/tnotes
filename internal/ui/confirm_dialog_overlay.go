package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
)

// ConfirmTarget は確認ダイアログの対象を表す。
type ConfirmTarget int

const (
	// ConfirmTargetNone は対象未指定。
	ConfirmTargetNone ConfirmTarget = iota
	// ConfirmTargetFolderDelete はフォルダ削除確認。
	ConfirmTargetFolderDelete
)

// ConfirmDialogOverlay は tui.ConfirmDialog をオーバーレイ化したラッパー。
type ConfirmDialogOverlay struct {
	dialog     *tui.ConfirmDialog
	target     ConfirmTarget
	folderName string
}

var _ shared.OverlayComponent = (*ConfirmDialogOverlay)(nil)

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

// Target は確認対象を返す。
func (c *ConfirmDialogOverlay) Target() ConfirmTarget { return c.target }

// FolderName は対象フォルダ名を返す。
func (c *ConfirmDialogOverlay) FolderName() string { return c.folderName }

func (c *ConfirmDialogOverlay) resultAction(result tui.ConfirmResult) ModelAction {
	switch result {
	case tui.ConfirmYes:
		return handleConfirmDialogResult(c.target, true, c.folderName)
	case tui.ConfirmNo:
		return handleConfirmDialogResult(c.target, false, c.folderName)
	case tui.ConfirmContinue:
		return nil
	}

	return nil
}

// --- ConfirmDialogOverlay → Model アクション ---

// handleConfirmDialogResult はオーバーレイをクリアし、Target に応じた後続処理を行う。
func handleConfirmDialogResult(target ConfirmTarget, confirmed bool, folderName string) ModelAction {
	return func(m *Model, ctx ActionContext) tea.Cmd {
		m.overlay = nil

		if target != ConfirmTargetFolderDelete || !confirmed {
			return nil
		}

		deleted, err := m.FolderList.DeleteFolder(folderName)
		if err != nil {
			return reportResult(err, "")(m, ctx)
		}

		info := "Deleted: " + folderName + " (" + strconv.Itoa(deleted) + " note(s) trashed)"

		return reportResult(nil, info)(m, ctx)
	}
}
