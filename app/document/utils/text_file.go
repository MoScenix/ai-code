package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/MoScenix/ai-code/app/document/conf"
	documentworkpool "github.com/MoScenix/ai-code/app/document/workpool"
	"github.com/MoScenix/ai-code/common/textsplitter"
	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	defaultParentChunkCount = 5
	defaultParentChunkStep  = 3
	defaultTaskChunkSize    = 200
	defaultEmbeddingBatch   = 25
	defaultWriteBatchSize   = 200
	maxEmbeddingBatch       = 25
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

	parentsDir := filepath.Join(filepath.Dir(textPath), "parents")
	if err := os.MkdirAll(parentsDir, 0o755); err != nil {
		return IndexTextFileResult{}, err
	}
	for _, parent := range result.Parents {
		name := fmt.Sprintf("%d.txt", parent.ID)
		if err := os.WriteFile(filepath.Join(parentsDir, name), []byte(parent.Content), 0o644); err != nil {
			return IndexTextFileResult{}, err
		}
	}

	if err := IndexChildChunks(ctx, result.Children); err != nil {
		return IndexTextFileResult{}, err
	}
	if err := refreshESIndex(ctx); err != nil {
		return IndexTextFileResult{}, err
	}

	klog.Infof("document index completed project_id=%d file_id=%d chunks=%d parents=%d", projectID, fileID, len(result.Children), len(result.Parents))
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

type indexTaskResult struct {
	index int
	err   error
}

type indexChildrenTask struct {
	ctx                context.Context
	index              int
	children           []ChildChunk
	embeddingBatchSize int
	writeBatchSize     int
	results            chan<- indexTaskResult
}

func (t *indexChildrenTask) Run(_ context.Context) error {
	err := indexChildBatch(t.ctx, t.children, t.embeddingBatchSize, t.writeBatchSize)
	t.results <- indexTaskResult{
		index: t.index,
		err:   err,
	}
	return err
}

func IndexChildChunks(ctx context.Context, children []ChildChunk) error {
	if len(children) == 0 {
		return nil
	}

	cfg := normalizedIndexConfig()
	batches := splitChildren(children, cfg.TaskChunkSize)
	pool, err := documentworkpool.Get()
	if err != nil {
		return err
	}

	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan indexTaskResult, len(batches))
	submitted := 0
	for i, batch := range batches {
		task := &indexChildrenTask{
			ctx:                taskCtx,
			index:              i,
			children:           batch,
			embeddingBatchSize: cfg.EmbeddingBatchSize,
			writeBatchSize:     cfg.WriteBatchSize,
			results:            results,
		}
		if err := pool.Submit(ctx, task); err != nil {
			cancel()
			return err
		}
		submitted++
	}

	var firstErr error
	for i := 0; i < submitted; i++ {
		select {
		case result := <-results:
			if result.err != nil && firstErr == nil {
				firstErr = fmt.Errorf("index child batch %d failed: %w", result.index, result.err)
				cancel()
			}
		case <-ctx.Done():
			cancel()
			return ctx.Err()
		}
	}
	return firstErr
}

func indexChildBatch(ctx context.Context, children []ChildChunk, embeddingBatchSize int, writeBatchSize int) error {
	if len(children) == 0 {
		return nil
	}
	vectors := make([][]float32, 0, len(children))
	for _, batch := range splitChildren(children, embeddingBatchSize) {
		texts := make([]string, 0, len(batch))
		for _, child := range batch {
			texts = append(texts, child.Content)
		}
		embedded, err := embedMilvusTexts(ctx, texts)
		if err != nil {
			return err
		}
		vectors = append(vectors, embedded...)
	}

	for _, span := range splitIndexSpans(len(children), writeBatchSize) {
		childBatch := children[span.start:span.end]
		vectorBatch := vectors[span.start:span.end]
		if err := insertESChildChunks(ctx, childBatch); err != nil {
			return err
		}
		if err := insertMilvusChildChunks(ctx, childBatch, vectorBatch); err != nil {
			return err
		}
	}
	return nil
}

type normalizedIndexSettings struct {
	TaskChunkSize      int
	EmbeddingBatchSize int
	WriteBatchSize     int
}

func normalizedIndexConfig() normalizedIndexSettings {
	cfg := conf.GetConf().Index
	taskChunkSize := cfg.TaskChunkSize
	if taskChunkSize <= 0 {
		taskChunkSize = defaultTaskChunkSize
	}
	embeddingBatchSize := cfg.EmbeddingBatchSize
	if embeddingBatchSize <= 0 {
		embeddingBatchSize = defaultEmbeddingBatch
	}
	if embeddingBatchSize > maxEmbeddingBatch {
		embeddingBatchSize = maxEmbeddingBatch
	}
	writeBatchSize := cfg.WriteBatchSize
	if writeBatchSize <= 0 {
		writeBatchSize = defaultWriteBatchSize
	}
	return normalizedIndexSettings{
		TaskChunkSize:      taskChunkSize,
		EmbeddingBatchSize: embeddingBatchSize,
		WriteBatchSize:     writeBatchSize,
	}
}

func splitChildren(children []ChildChunk, size int) [][]ChildChunk {
	if size <= 0 {
		size = len(children)
	}
	batches := make([][]ChildChunk, 0, (len(children)+size-1)/size)
	for start := 0; start < len(children); start += size {
		end := start + size
		if end > len(children) {
			end = len(children)
		}
		batches = append(batches, children[start:end])
	}
	return batches
}

type indexSpan struct {
	start int
	end   int
}

func splitIndexSpans(total int, size int) []indexSpan {
	if size <= 0 {
		size = total
	}
	spans := make([]indexSpan, 0, (total+size-1)/size)
	for start := 0; start < total; start += size {
		end := start + size
		if end > total {
			end = total
		}
		spans = append(spans, indexSpan{start: start, end: end})
	}
	return spans
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
