// Package pdfx extracts plain text from a PDF using a pure-Go library, so there
// are no system dependencies to install. Note: scanned/image-only PDFs contain
// no text layer and will return empty output (they need OCR first).
package pdfx

import (
	"bytes"
	"fmt"

	"github.com/ledongthuc/pdf"
)

// ExtractText returns all readable text from the PDF at path.
func ExtractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	rc, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extract text: %w", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		return "", fmt.Errorf("read text: %w", err)
	}
	return buf.String(), nil
}
