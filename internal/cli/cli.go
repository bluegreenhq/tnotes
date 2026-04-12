package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
)

const cmdName = "tnotes"

var ErrUnknownCommand = errors.New("unknown command")

// Version はビルド時に注入されるバージョン文字列。未設定時は "dev"。
var Version string //nolint:gochecknoglobals // ビルド時にmain.goから注入される

const minArgsForSubcommand = 2

// Run はCLIサブコマンドを実行する。
// サブコマンドが指定されていない場合は false を返し、TUIにフォールバックさせる。
// サブコマンドが指定されている場合は true を返す。
func Run(args []string, a *app.App, r io.Reader, w io.Writer) (bool, error) { //nolint:cyclop // サブコマンドのディスパッチで増加する
	if len(args) < minArgsForSubcommand {
		return false, nil
	}

	switch args[1] {
	case "folder":
		return true, runFolder(args, a, r, w)
	case "move":
		return true, runMove(args, a, w)
	case "list":
		return true, runList(args, a, w)
	case "purge":
		return true, runPurge(args, a, r, w)
	case "get":
		return true, runGet(args, a, w)
	case "create":
		return true, runCreate(args, a, r, w)
	case "update":
		return true, runUpdate(args, a, r, w)
	case "delete":
		return true, runDelete(args, a, w)
	case "search":
		return true, runSearch(args, a, w)
	case "export":
		return true, runExport(args, a, w)
	case "import":
		return true, runImport(args, a, w)
	case "version":
		return true, runVersion(w)
	case "help":
		printUsage(w)

		return true, nil
	default:
		_, _ = fmt.Fprintf(w, "unknown command: %s\n\n", args[1])
		printUsage(w)

		return true, errors.WithDetail(ErrUnknownCommand, args[1])
	}
}

func printUsage(w io.Writer) {
	commands := []struct{ cmd, desc string }{
		{"", "TUIモードで起動"},
		{"--no-wrap", "TUIモードで起動（水平スクロールモード）"},
		{"list [--json]", "ノート一覧を表示"},
		{"list --trash", "ゴミ箱のノート一覧を表示"},
		{"list --folder <name>", "指定フォルダのノート一覧を表示"},
		{"search <query> [--folder <name>] [--json]", "全文検索"},
		{"get <id> [--json]", "指定IDのノートを表示（ゴミ箱含む）"},
		{"create [file] [--folder <name>]", "ファイルまたは標準入力からノートを作成"},
		{"update <id> [file]", "ノートの本文を上書き更新"},
		{"delete <id>", "ノートをゴミ箱に移動"},
		{"move <id> <folder>", "ノートを指定フォルダに移動"},
		{"purge", "ゴミ箱を空にする（確認あり）"},
		{"purge --force", "ゴミ箱を空にする（確認なし）"},
		{"export <file>", "データ一式をzipにエクスポート"},
		{"import <file>", "zipからデータをインポート"},
		{"version", "バージョンを表示"},
		{"folder list [--json]", "フォルダ一覧を表示"},
		{"folder create <name>", "フォルダを作成"},
		{"folder rename <old> <new>", "フォルダをリネーム"},
		{"folder delete <name>", "フォルダを削除"},
		{"help", "このヘルプを表示"},
	}

	_, _ = fmt.Fprintln(w, "Usage:")

	for _, c := range commands {
		_, _ = fmt.Fprintf(w, "  %s\n", usageLine(c.cmd, c.desc))
	}
}

func usageLine(cmd, desc string) string {
	full := cmdName
	if cmd != "" {
		full += " " + cmd
	}

	const minPadding = 2

	maxLen := len(cmdName) + 1 + len("search <query> [--folder <name>] [--json]")
	pad := max(maxLen-len(full)+minPadding, minPadding)

	return full + strings.Repeat(" ", pad) + desc
}

func runVersion(w io.Writer) error {
	v := Version
	if v == "" {
		v = "dev"
	}

	_, err := fmt.Fprintf(w, "tnotes version %s\n", v)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
