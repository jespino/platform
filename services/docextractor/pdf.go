package docextractor

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ledongthuc/pdf"
)

type pdfExtractor struct{}

func (pe *pdfExtractor) Match(filename string) bool {
	supportedExtensions := map[string]bool{
		"pdf": true,
	}
	extension := strings.TrimPrefix(path.Ext(filename), ".")
	if supportedExtensions[extension] {
		return true
	}
	return false
}

func (pe *pdfExtractor) Extract(filename string, r io.Reader) (string, error) {
	f, err := ioutil.TempFile(os.TempDir(), "pdflib")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())
	size, err := io.Copy(f, r)
	if err != nil {
		return "", fmt.Errorf("error copying data into temporary file: %v", err)
	}

	reader, err := pdf.NewReader(f, size)
	if err != nil {
		return "", err
	}

	text := ""
	totalPage := reader.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := reader.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			for _, word := range row.Content {
				text += " " + fmt.Sprintf(word.S)
			}
		}
	}

	var buf bytes.Buffer
	b, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}
