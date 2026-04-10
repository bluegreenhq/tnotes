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
	Title      string
	Detail     string
	confirmBtn string
	cancelBtn  string
	hoverYes   bool
	hoverNo    bool
}

// NewConfirmDialog は新しい ConfirmDialog を生成する。
func NewConfirmDialog(title, detail string) ConfirmDialog {
	return ConfirmDialog{
		Title:      title,
		Detail:     detail,
		confirmBtn: "[Y]es",
		cancelBtn:  "[N]o",
		hoverYes:   false,
		hoverNo:    false,
	}
}

// HoverYes は確定ボタンがホバー中かを返す。
func (d *ConfirmDialog) HoverYes() bool { return d.hoverYes }

// HoverNo はキャンセルボタンがホバー中かを返す。
func (d *ConfirmDialog) HoverNo() bool { return d.hoverNo }

const (
	confirmContentLines = 4 // title + detail + empty + buttons
	confirmBorderPad    = 4 // border上1 + padding上1 + padding下1 + border下1
)

// contentLines はダイアログコンテンツ（padding内側）の行数を返す。
func (d *ConfirmDialog) contentLines() int {
	return confirmContentLines
}
