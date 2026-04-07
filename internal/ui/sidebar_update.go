package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Update はメッセージに応じてサイドバーの状態を更新する。
// trashMode はゴミ箱モードかどうかを示す。
// ナビゲーション（カーソル移動）は自身で処理し、
// ノート操作（作成、削除等）は tea.Cmd で SidebarMsg を返して Model に委譲する。
func (s *Sidebar) Update(msg tea.Msg, now time.Time, trashMode bool) (Sidebar, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return *s, nil
	}

	if keyMsg.Mod&tea.ModCtrl != 0 {
		return s.handleCtrlKey(keyMsg, now)
	}

	if trashMode {
		return s.handleTrashModeKey(keyMsg, now)
	}

	return s.handleNormalKey(keyMsg, now)
}

func (s *Sidebar) handleNormalKey(msg tea.KeyPressMsg, now time.Time) (Sidebar, tea.Cmd) {
	switch msg.Code {
	case 'n':
		return *s, sidebarCmd(SidebarCreate)
	case 'd', tea.KeyDelete, tea.KeyBackspace:
		return *s, sidebarCmd(SidebarTrash)
	case 'g':
		return *s, sidebarCmd(SidebarEnterTrash)
	case tea.KeyUp, 'k':
		return s.moveUpCmd(now)
	case tea.KeyDown, 'j':
		return s.moveDownCmd(now)
	case tea.KeyEnter:
		return *s, sidebarCmd(SidebarEdit)
	}

	return *s, nil
}

func (s *Sidebar) handleTrashModeKey(msg tea.KeyPressMsg, now time.Time) (Sidebar, tea.Cmd) {
	switch msg.Code {
	case 'g':
		return *s, sidebarCmd(SidebarExitTrash)
	case 'r':
		return *s, sidebarCmd(SidebarRestore)
	case tea.KeyUp, 'k':
		return s.moveUpCmd(now)
	case tea.KeyDown, 'j':
		return s.moveDownCmd(now)
	}

	return *s, nil
}

func (s *Sidebar) handleCtrlKey(msg tea.KeyPressMsg, now time.Time) (Sidebar, tea.Cmd) {
	switch msg.Code {
	case 'z':
		if msg.Mod&tea.ModShift != 0 {
			return *s, sidebarCmd(SidebarRedo)
		}

		return *s, sidebarCmd(SidebarUndo)
	case 'n':
		return s.moveDownCmd(now)
	case 'p':
		return s.moveUpCmd(now)
	}

	return *s, nil
}

func (s *Sidebar) moveUpCmd(now time.Time) (Sidebar, tea.Cmd) {
	prev := s.selected
	s.MoveUp(now)

	if s.selected != prev {
		return *s, sidebarCmd(SidebarSelect)
	}

	return *s, nil
}

func (s *Sidebar) moveDownCmd(now time.Time) (Sidebar, tea.Cmd) {
	prev := s.selected
	s.MoveDown(now)

	if s.selected != prev {
		return *s, sidebarCmd(SidebarSelect)
	}

	return *s, nil
}

func sidebarCmd(msg SidebarMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}
