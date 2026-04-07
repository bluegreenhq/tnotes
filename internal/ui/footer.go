package ui

// HoverTarget はフッターボタンのホバーターゲット。
type HoverTarget int

const (
	HoverNone HoverTarget = iota
	HoverNew
	HoverQuit
	HoverRestore
	HoverTrashToggle
	HoverCopy
	HoverCut
)

// FooterButton はフッターに表示する1つのボタンを表す。
type FooterButton struct {
	Label    string
	Target   HoverTarget
	Disabled bool
}

// NewFooterButton は新しい FooterButton を生成する。
func NewFooterButton(label string, target HoverTarget) FooterButton {
	return FooterButton{Label: label, Target: target, Disabled: false}
}

// Footer はフッターバーの状態を表す。
type Footer struct {
	hover   HoverTarget
	buttons []FooterButton
}

// Hover はホバーターゲットを返す。
func (f *Footer) Hover() HoverTarget { return f.hover }

// FooterState はフッターのボタン構築に必要な状態。
type FooterState struct {
	TrashMode    bool
	TrashCount   int
	HasSelection bool
	EditorDirty  bool
}

// RebuildButtons はフッターのボタンリストを再構築する。
func (f *Footer) RebuildButtons(s FooterState) {
	var buttons []FooterButton

	if s.TrashMode {
		buttons = []FooterButton{
			{Label: "[Restore]", Target: HoverRestore, Disabled: s.TrashCount == 0},
			NewFooterButton("[Notes]", HoverTrashToggle),
			NewFooterButton("[Quit]", HoverQuit),
		}
	} else {
		buttons = []FooterButton{
			NewFooterButton("[New]", HoverNew),
			NewFooterButton("[Trash]", HoverTrashToggle),
			NewFooterButton("[Quit]", HoverQuit),
		}

		if s.HasSelection {
			buttons = append(buttons,
				NewFooterButton("[Copy]", HoverCopy),
				NewFooterButton("[Cut]", HoverCut),
			)
		} else if s.EditorDirty {
			buttons = append(buttons,
				FooterButton{Label: "● Modified", Target: HoverNone, Disabled: true},
			)
		}
	}

	f.buttons = buttons
}
