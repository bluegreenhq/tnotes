package ui

// MenuItem はポップアップメニューの1項目を表す。
type MenuItem struct {
	Label    string
	Disabled bool
}

// PopupMenu は汎用ポップアップメニューコンポーネント。
type PopupMenu struct {
	items []MenuItem
	hover int // ホバー中の項目インデックス (-1 = なし)
}

// NewPopupMenu は新しい PopupMenu を生成する。
func NewPopupMenu(items []MenuItem) *PopupMenu {
	return &PopupMenu{items: items, hover: -1}
}

// Items はメニュー項目のリストを返す。
func (m *PopupMenu) Items() []MenuItem { return m.items }

const (
	menuBorderLines = 2 // 上枠 + 下枠
	menuPadding     = 4 // │ + space + space + │
)

// Height はメニュー全体の高さ（上枠 + 項目数*2-1(項目間空行) + 下枠）を返す。
func (m *PopupMenu) Height() int {
	n := len(m.items)
	if n == 0 {
		return 0
	}

	// 上枠1 + (項目数 + 項目間空行(n-1)) + 下枠1
	return menuBorderLines + n + n - 1
}

// Width はメニュー全体の幅（左枠 + スペース + 最長ラベル + スペース + 右枠）を返す。
func (m *PopupMenu) Width() int {
	maxLen := 0
	for _, item := range m.items {
		if len(item.Label) > maxLen {
			maxLen = len(item.Label)
		}
	}

	return maxLen + menuPadding
}

// Hover はホバー中の項目インデックスを返す。-1 はホバーなし。
func (m *PopupMenu) Hover() int { return m.hover }
