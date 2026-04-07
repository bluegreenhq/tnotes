package ui_test

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

type screenBuf struct {
	grid   [][]byte
	row    int
	col    int
	w, h   int
	parser *ansi.Parser
	state  byte
}

func newScreenBuf(w, h int) *screenBuf {
	grid := make([][]byte, h)
	for i := range grid {
		grid[i] = make([]byte, w)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	return &screenBuf{grid: grid, w: w, h: h, parser: ansi.NewParser()}
}

func (s *screenBuf) String() string {
	var sb strings.Builder

	for _, line := range s.grid {
		sb.Write(line)
		sb.WriteByte('\n')
	}

	return sb.String()
}

func (s *screenBuf) inBounds() bool {
	return s.row >= 0 && s.row < s.h && s.col >= 0 && s.col < s.w
}

func (s *screenBuf) putChar(seq []byte, width int) {
	if s.inBounds() {
		if len(seq) == 1 {
			s.grid[s.row][s.col] = seq[0]
		} else {
			s.grid[s.row][s.col] = '?'
		}
	}

	s.col += width
}

func (s *screenBuf) handleCSI(cmd ansi.Cmd) {
	switch cmd.Final() {
	case 'H':
		r, _ := s.parser.Param(0, 1)
		c, _ := s.parser.Param(1, 1)
		s.row = r - 1
		s.col = c - 1
	case 'K':
		param, _ := s.parser.Param(0, 0)
		if param == 0 && s.inBounds() {
			for j := s.col; j < s.w; j++ {
				s.grid[s.row][j] = ' '
			}
		}
	case 'J':
		param, _ := s.parser.Param(0, 0)
		if param == 2 {
			for i := range s.grid {
				for j := range s.grid[i] {
					s.grid[i][j] = ' '
				}
			}
		}
	}
}

// renderScreen は生のターミナル出力を簡易画面バッファに描画し、テキストを返す。
// ansi.Strip と異なり、CUP（カーソル位置移動）を正しく処理する。
func renderScreen(bts []byte, w, h int) string {
	buf := newScreenBuf(w, h)

	for len(bts) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(bts, buf.state, buf.parser)
		buf.state = newState

		switch {
		case width > 0:
			buf.putChar(seq, width)
		case len(seq) > 0 && seq[0] == '\n':
			buf.row++
			buf.col = 0
		case len(seq) > 0 && seq[0] == '\r':
			buf.col = 0
		case len(seq) > 1 && seq[0] == '\x1b':
			buf.handleCSI(ansi.Cmd(buf.parser.Command()))
		}

		bts = bts[n:]
	}

	return buf.String()
}
