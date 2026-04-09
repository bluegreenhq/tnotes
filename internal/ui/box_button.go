package ui

// boxButtonPadding は罫線ボタンの追加幅（│ + space + space + │）。
const boxButtonPadding = 4

// BoxButton は罫線で囲まれたボタンコンポーネント。
type BoxButton struct {
	label   string
	hovered bool
}

// NewBoxButton は新しい BoxButton を生成する。
func NewBoxButton(label string) BoxButton {
	return BoxButton{label: label, hovered: false}
}

// Label はボタンのラベルを返す。
func (b *BoxButton) Label() string { return b.label }

// Hovered はホバー状態を返す。
func (b *BoxButton) Hovered() bool { return b.hovered }

// SetHovered はホバー状態を設定する。
func (b *BoxButton) SetHovered(h bool) { b.hovered = h }

// DisplayWidth はボタンの表示幅（│ + space + label + space + │）を返す。
func (b *BoxButton) DisplayWidth() int {
	return len(b.label) + boxButtonPadding
}

// innerLabel は罫線内の文字列（スペース含む）を返す。
func (b *BoxButton) innerLabel() string {
	return " " + b.label + " "
}
