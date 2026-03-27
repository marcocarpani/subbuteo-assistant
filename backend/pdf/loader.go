package pdf

import (
	"bytes"
	"fmt"

	"github.com/ledongthuc/pdf"
)

// ExtractText estrae tutto il testo dal PDF in memoria
func ExtractText(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("apertura PDF: %w", err)
	}

	var buf bytes.Buffer
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue // Salta pagine problematiche (es. immagini)
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
