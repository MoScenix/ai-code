package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	docutils "github.com/MoScenix/ai-code/app/document/utils"
	"github.com/MoScenix/ai-code/common/filestore"
	"github.com/ledongthuc/pdf"
)

func projectFileDir(projectID int64, fileID int64) string {
	shareDir := filepath.Clean(filestore.GetConf().ShareDir.ShareDir)
	staticDir := filepath.Dir(shareDir)
	return filepath.Join(staticDir, "document", strconv.FormatInt(projectID, 10), strconv.FormatInt(fileID, 10))
}

func findFileByExt(dir string, ext string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	ext = strings.ToLower(ext)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.EqualFold(filepath.Ext(name), ext) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "", fmt.Errorf("document: no %s file in %s", ext, dir)
	}
	sort.Strings(names)
	return filepath.Join(dir, names[0]), nil
}

func parsePDFToTextFile(pdfPath string) (string, int64, error) {
	file, reader, err := pdf.Open(pdfPath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", 0, err
	}
	raw, err := io.ReadAll(textReader)
	if err != nil {
		return "", 0, err
	}

	text := docutils.CleanText(string(raw))
	txtPath := strings.TrimSuffix(pdfPath, filepath.Ext(pdfPath)) + ".txt"
	if err := os.WriteFile(txtPath, []byte(text), 0o644); err != nil {
		return "", 0, err
	}
	return txtPath, int64(len([]byte(text))), nil
}
