package cli

import (
	"fmt"
	"io"

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
	case "list":
		return true, runList(args, a, w)
	case "purge":
		return true, runPurge(args, a, r, w)
	case "get":
		return true, runGet(args, a, w)
	case "create":
		return true, runCreate(args, a, r, w)
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
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintf(w, "  %s              TUIモードで起動\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s --no-wrap    TUIモードで起動（水平スクロールモード）\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s list           ノート一覧を表示\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s list --trash   ゴミ箱のノート一覧を表示\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s purge          ゴミ箱を空にする（確認あり）\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s purge --force   ゴミ箱を空にする（確認なし）\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s get <id>       指定IDのノートを表示（ゴミ箱含む）\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s create [file] ファイルまたは標準入力からノートを作成\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s export <file> データ一式をzipにエクスポート\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s import <file> zipからデータをインポート\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s version       バージョンを表示\n", cmdName)
	_, _ = fmt.Fprintf(w, "  %s help         このヘルプを表示\n", cmdName)
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
