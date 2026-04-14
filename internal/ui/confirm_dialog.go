package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// ConfirmResult はダイアログの操作結果を表す。
type ConfirmResult int

const (
	// ConfirmContinue はダイアログ継続中。
	ConfirmContinue ConfirmResult = iota
	// ConfirmYes は確定。
	ConfirmYes
	// ConfirmNo はキャンセル。
	ConfirmNo
)

// ConfirmDialog は汎用確認ダイアログコンポーネント。
type ConfirmDialog struct {
	Title       string
	Detail      string
	yesBtn      BoxButton
	noBtn       BoxButton
	screenWidth int // 画面幅
	bodyHeight  int // ボディ領域の高さ
}

// NewConfirmDialog は新しい ConfirmDialog を生成する。
func NewConfirmDialog(title, detail string) ConfirmDialog {
	return ConfirmDialog{
		Title:       title,
		Detail:      detail,
		yesBtn:      NewBoxButton("[Y]es"),
		noBtn:       NewBoxButton("[N]o"),
		screenWidth: 0,
		bodyHeight:  0,
	}
}

const (
	confirmContentLines  = 7 // title + empty + detail + empty + button top + button middle + button bottom
	confirmBorderPad     = 3 // border上1 + padding上1 + border下1
	confirmButtonGapCols = 2 // ボタン間のスペース数
	confirmButtonRows    = 3 // ボタン描画行数（上枠・ラベル・下枠）
	confirmPaddingSides  = 2 // 左右パディングの個数
	confirmCenterDiv     = 2 // センタリング用除数
)

// SetScreenSize は画面サイズを設定する。
func (d *ConfirmDialog) SetScreenSize(screenWidth, bodyHeight int) {
	d.screenWidth = screenWidth
	d.bodyHeight = bodyHeight
}

// Origin はダイアログの画面上のコンテンツ左上座標を返す。
// border + padding 分を加算した座標。
func (d *ConfirmDialog) Origin() (int, int) {
	rendered := d.View()
	dialogLines := strings.Split(rendered, "\n")

	const (
		borderPaddingLines = 2 // border上 + padding上
		borderPaddingCols  = 3 // border左1 + padding左2
	)

	dialogWidth := lipgloss.Width(dialogLines[0])

	sy := max((d.bodyHeight-len(dialogLines))/confirmCenterDiv, 0) + borderPaddingLines
	sx := max((d.screenWidth-dialogWidth)/confirmCenterDiv, 0) + borderPaddingCols

	return sx, sy
}

// buttonPadLeft はボタン領域のセンタリング用左パディングを返す。
func (d *ConfirmDialog) buttonPadLeft() int {
	buttonsWidth := d.yesBtn.DisplayWidth() + confirmButtonGapCols + d.noBtn.DisplayWidth()
	contentWidth := confirmDialogWidth - confirmDialogPaddingH*confirmPaddingSides

	return (contentWidth - buttonsWidth) / confirmCenterDiv
}

// contentLines はダイアログコンテンツ（padding内側）の行数を返す。
func (d *ConfirmDialog) contentLines() int {
	return confirmContentLines
}
