# VHS demo recording for Notes TUI
# Usage: vhs demo.tape

Set Width 900
Set Height 500
Set FontSize 14
Set Shell "bash"

Hide
Type "printf '貼り付け用テキスト' | pbcopy"
Enter
Sleep 500ms
Type "rm -rf /tmp/tnotes-demo && NOTES_DATA_DIR=/tmp/tnotes-demo faketime '2026-01-01 12:00:00' go run ."
Enter
Sleep 2s
Show

# ノート作成（日本語入力）
Type "n"
Sleep 1s
Type "日本語メモ"
Enter
Type "これはテスト用のメモです。"
Enter
Type "範囲選択のデモを行います。"
Sleep 2s

# サイドバーに戻る
Escape
Sleep 1s

# 新しいノートを作成して貼り付け
Type "n"
Sleep 1s
Type "貼り付けテスト"
Enter
Ctrl+V
Sleep 2s

# サイドバーに戻って確認
Escape
Sleep 1s
Type "j"
Sleep 1s
Tab
Sleep 2s

# 終了
Escape
Sleep 500ms
Type "q"
Sleep 500ms
