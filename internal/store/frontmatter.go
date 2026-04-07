package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"

	"github.com/bluegreenhq/tnotes/internal/note"
)

var (
	ErrMissingFrontmatter        = errors.New("missing frontmatter delimiter")
	ErrMissingClosingFrontmatter = errors.New("missing closing frontmatter delimiter")
)

const frontmatterDelimiter = "---"

type frontmatter struct {
	ID        note.NoteID `yaml:"id"`
	CreatedAt time.Time   `yaml:"created_at"`
	UpdatedAt time.Time   `yaml:"updated_at"`
}

// marshalNoteFile はNoteをフロントマター付きMarkdown文字列に変換する。
func marshalNoteFile(n note.Note) string {
	fm := frontmatter{
		ID:        n.ID,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal frontmatter: %v", err))
	}

	var b strings.Builder
	b.WriteString(frontmatterDelimiter)
	b.WriteString("\n")
	b.Write(yamlBytes)
	b.WriteString(frontmatterDelimiter)
	b.WriteString("\n")
	b.WriteString(n.Body)

	return b.String()
}

// parseNoteFile はフロントマター付きMarkdown文字列をNoteに変換する。
func parseNoteFile(content string) (note.Note, error) {
	if !strings.HasPrefix(content, frontmatterDelimiter+"\n") {
		return note.Note{}, ErrMissingFrontmatter
	}

	rest := content[len(frontmatterDelimiter)+1:]

	endIdx := strings.Index(rest, frontmatterDelimiter+"\n")
	if endIdx < 0 {
		return note.Note{}, ErrMissingClosingFrontmatter
	}

	yamlStr := rest[:endIdx]
	body := rest[endIdx+len(frontmatterDelimiter)+1:]

	var fm frontmatter

	err := yaml.Unmarshal([]byte(yamlStr), &fm)
	if err != nil {
		return note.Note{}, errors.WithStack(err)
	}

	return note.Note{
		Metadata: note.Metadata{
			ID:        fm.ID,
			Title:     "",
			CreatedAt: fm.CreatedAt,
			UpdatedAt: fm.UpdatedAt,
			Path:      "",
		},
		Body: body,
	}, nil
}
