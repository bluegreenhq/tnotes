package ui

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
	Title  string
	Detail string
	yesBtn BoxButton
	noBtn  BoxButton
}

// NewConfirmDialog は新しい ConfirmDialog を生成する。
func NewConfirmDialog(title, detail string) ConfirmDialog {
	return ConfirmDialog{
		Title:  title,
		Detail: detail,
		yesBtn: NewBoxButton("[Y]es"),
		noBtn:  NewBoxButton("[N]o"),
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
