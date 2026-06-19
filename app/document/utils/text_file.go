package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/MoScenix/ai-code/common/textsplitter"
)

const (
	defaultParentChunkCount = 5
	defaultParentChunkStep  = 3
)

type ChildChunkMeta struct {
	ProjectID int64   `json:"projectId"`
	FileID    int64   `json:"fileId"`
	ChunkID   int64   `json:"chunkId"`
	ParentIDs []int64 `json:"parentIds"`
}

type ChildChunk struct {
	Meta    ChildChunkMeta
	Content string
}

type ParentChunk struct {
	ID      int64
	Content string
}

type SplitResult struct {
	Children []ChildChunk
	Parents  []ParentChunk
}

type IndexTextFileResult struct {
	ChunkCount  int64
	ParentCount int64
}

func IndexTextFile(ctx context.Context, projectID int64, fileID int64, textPath string, minSize int64, maxSize int64) (IndexTextFileResult, error) {
	raw, err := os.ReadFile(textPath)
	if err != nil {
		return IndexTextFileResult{}, err
	}

	result := SplitTextWithParents(projectID, fileID, CleanText(string(raw)), minSize, maxSize)
	if len(result.Children) == 0 {
		return IndexTextFileResult{}, nil
	}

	chunksDir := filepath.Join(filepath.Dir(textPath), "chunks")
	if err := os.MkdirAll(chunksDir, 0o755); err != nil {
		return IndexTextFileResult{}, err
	}
	for _, parent := range result.Parents {
		name := fmt.Sprintf("parent_%d.txt", parent.ID)
		if err := os.WriteFile(filepath.Join(chunksDir, name), []byte(parent.Content), 0o644); err != nil {
			return IndexTextFileResult{}, err
		}
	}

	for _, child := range result.Children {
		if err := InsertChildChunk(ctx, child.Meta, child.Content); err != nil {
			return IndexTextFileResult{}, err
		}
	}

	return IndexTextFileResult{
		ChunkCount:  int64(len(result.Children)),
		ParentCount: int64(len(result.Parents)),
	}, nil
}

func SplitTextWithParents(projectID int64, fileID int64, text string, minSize int64, maxSize int64) SplitResult {
	chunks := textsplitter.SplitAll(text, textsplitter.Options{
		MinSize: int(minSize),
		MaxSize: int(maxSize),
	})
	if len(chunks) == 0 {
		return SplitResult{}
	}

	parentIDsByChild, parents := buildParents(chunks)
	children := make([]ChildChunk, 0, len(chunks))
	for i, chunk := range chunks {
		children = append(children, ChildChunk{
			Meta: ChildChunkMeta{
				ProjectID: projectID,
				FileID:    fileID,
				ChunkID:   int64(i + 1),
				ParentIDs: parentIDsByChild[i],
			},
			Content: chunk.Text,
		})
	}

	return SplitResult{
		Children: children,
		Parents:  parents,
	}
}

func CleanText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var builder strings.Builder
	builder.Grow(len(text))
	for _, r := range text {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}

	lines := strings.Split(builder.String(), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRightFunc(line, unicode.IsSpace)
	}
	text = strings.Join(lines, "\n")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func InsertChildChunk(ctx context.Context, meta ChildChunkMeta, content string) error {
	if err := insertESChildChunk(ctx, meta, content); err != nil {
		return err
	}
	return insertMilvusChildChunk(ctx, meta, content)
}

func buildParents(chunks []textsplitter.Chunk) ([][]int64, []ParentChunk) {
	parentIDsByChild := make([][]int64, len(chunks))
	parents := make([]ParentChunk, 0, len(chunks)/defaultParentChunkStep+1)
	parentID := int64(1)

	for start := 0; start < len(chunks); start += defaultParentChunkStep {
		end := start + defaultParentChunkCount
		if end > len(chunks) {
			end = len(chunks)
		}
		if start >= end {
			break
		}

		parts := make([]string, 0, end-start)
		for i := start; i < end; i++ {
			parts = append(parts, strings.TrimSpace(chunks[i].Text))
			parentIDsByChild[i] = append(parentIDsByChild[i], parentID)
		}
		parents = append(parents, ParentChunk{
			ID:      parentID,
			Content: strings.Join(parts, "\n\n"),
		})
		parentID++

		if end == len(chunks) {
			break
		}
	}

	return parentIDsByChild, parents
}
