package service

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	var selected os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.EqualFold(filepath.Ext(name), ext) {
			if selected == nil {
				selected = entry
				continue
			}
			selectedInfo, selectedErr := selected.Info()
			entryInfo, entryErr := entry.Info()
			if selectedErr == nil && entryErr == nil && entryInfo.ModTime().After(selectedInfo.ModTime()) {
				selected = entry
			}
		}
	}
	if selected == nil {
		return "", fmt.Errorf("document: no %s file in %s", ext, dir)
	}
	return filepath.Join(dir, selected.Name()), nil
}

func parsePDFToTextFile(pdfPath string) (string, int64, error) {
	if txtPath, size, err := parsePDFToTextFileWithPoppler(pdfPath); err == nil {
		return txtPath, size, nil
	}
	return parsePDFToTextFileWithGoPDF(pdfPath)
}

func parsePDFToTextFileWithPoppler(pdfPath string) (string, int64, error) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", 0, err
	}

	txtPath := strings.TrimSuffix(pdfPath, filepath.Ext(pdfPath)) + ".txt"
	cmd := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", pdfPath, txtPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", 0, fmt.Errorf("pdftotext failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	raw, err := os.ReadFile(txtPath)
	if err != nil {
		return "", 0, err
	}
	text := docutils.CleanText(string(raw))
	if err := os.WriteFile(txtPath, []byte(text), 0o644); err != nil {
		return "", 0, err
	}
	return txtPath, int64(len([]byte(text))), nil
}

func parsePDFToTextFileWithGoPDF(pdfPath string) (string, int64, error) {
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
