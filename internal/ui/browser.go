package ui

import (
	"context"
	"os/exec"
	"runtime"

	tea "charm.land/bubbletea/v2"
)

// openURLInBrowser は URL をデフォルトブラウザで開く tea.Cmd を返す。
func openURLInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.CommandContext(ctx, "open", url)
		case "windows":
			cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "", url)
		default: // linux, freebsd, etc.
			cmd = exec.CommandContext(ctx, "xdg-open", url)
		}

		_ = cmd.Start()

		return nil
	}
}
