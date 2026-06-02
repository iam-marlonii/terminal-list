// Package document reads text from supported import file types.
package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iam-marlonjr/terminal-list/internal/pdfx"
)

// ExtractText returns plain text from a PDF, Markdown, or text file.
func ExtractText(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return pdfx.ExtractText(path)
	case ".md", ".markdown", ".txt", ".text":
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return "", fmt.Errorf("unsupported file type %q (use .pdf, .md, or .txt)", ext)
	}
}
