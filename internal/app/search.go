package app

import (
	"strings"

	"github.com/bluegreenhq/tnotes/internal/note"
)

// SearchByFolder は指定フォルダ内のノートをキーワード検索して返す。
// query が空の場合は全ノートを返す。大文字小文字を区別しない。
func (a *App) SearchByFolder(folderName string, query string) []note.Note {
	notes := a.ListByFolder(folderName)
	if query == "" {
		return notes
	}

	q := strings.ToLower(query)
	results := make([]note.Note, 0)

	for _, n := range notes {
		loaded, err := a.LoadNote(n)
		if err != nil {
			continue
		}

		if matchesQuery(loaded, q) {
			results = append(results, loaded)
		}
	}

	return results
}

// ExtractSnippets は本文からクエリにマッチする箇所を前後 contextSize 文字のスニペットとして返す。
func ExtractSnippets(body string, query string, contextSize int) []string {
	lowerQuery := strings.ToLower(query)
	lower := strings.ToLower(body)
	runes := []rune(body)
	lowerRunes := []rune(lower)
	queryRunes := []rune(lowerQuery)
	queryLen := len(queryRunes)

	var snippets []string

	for i := 0; i <= len(lowerRunes)-queryLen; i++ {
		if string(lowerRunes[i:i+queryLen]) != lowerQuery {
			continue
		}

		start := max(i-contextSize, 0)
		end := min(i+queryLen+contextSize, len(runes))

		var sb strings.Builder

		if start > 0 {
			sb.WriteString("...")
		}

		sb.WriteString(string(runes[start:end]))

		if end < len(runes) {
			sb.WriteString("...")
		}

		snippets = append(snippets, collapseWhitespace(sb.String()))

		i += queryLen - 1
	}

	return snippets
}

func matchesQuery(n note.Note, lowerQuery string) bool {
	title := strings.ToLower(n.Title())
	if strings.Contains(title, lowerQuery) {
		return true
	}

	body := strings.ToLower(n.Body)

	return strings.Contains(body, lowerQuery)
}

func collapseWhitespace(s string) string {
	var sb strings.Builder

	prevSpace := false

	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			if !prevSpace {
				sb.WriteByte(' ')
			}

			prevSpace = true

			continue
		}

		sb.WriteRune(r)

		prevSpace = false
	}

	return sb.String()
}
