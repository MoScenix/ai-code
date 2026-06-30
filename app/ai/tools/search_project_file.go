package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MoScenix/ai-code/app/ai/conf"
	"github.com/MoScenix/ai-code/app/ai/infra"
	aiutils "github.com/MoScenix/ai-code/app/ai/utils"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type SearchProjectFileParams struct {
	FileID int64  `json:"file_id" jsonschema:"description=要搜索的上传文件 ID"`
	Query  string `json:"query" jsonschema:"description=搜索内容，使用自然语言描述要查找的资料"`
	TopK   int64  `json:"top_k,omitempty" jsonschema:"description=返回父块数量，默认 5"`
}

type SearchProjectFileResult struct {
	Parents []SearchProjectFileParent `json:"parents"`
	Error   string                    `json:"error,omitempty"`
}

type SearchProjectFileParent struct {
	ParentID int64  `json:"parent_id"`
	Content  string `json:"content"`
}

func SearchProjectFileFunc(ctx context.Context, params *SearchProjectFileParams) (SearchProjectFileResult, error) {
	if params == nil {
		return SearchProjectFileResult{Error: "missing params"}, nil
	}
	if params.FileID <= 0 {
		return SearchProjectFileResult{Error: "file_id must be positive"}, nil
	}
	query := strings.TrimSpace(params.Query)
	if query == "" {
		return SearchProjectFileResult{Error: "query cannot be empty"}, nil
	}

	projectIDRaw, ok := aiutils.ProjectIDFromContext(ctx)
	if !ok || strings.TrimSpace(projectIDRaw) == "" {
		return SearchProjectFileResult{Error: "missing project id in context"}, nil
	}
	projectID, err := strconv.ParseInt(projectIDRaw, 10, 64)
	if err != nil {
		return SearchProjectFileResult{Error: fmt.Sprintf("invalid project id: %s", projectIDRaw)}, nil
	}

	client, err := infra.DocumentClient()
	if err != nil {
		return SearchProjectFileResult{Error: err.Error()}, nil
	}
	resp, err := client.SearchFile(ctx, &document.SearchFileReq{
		ProjectId: projectID,
		FileId:    params.FileID,
		Query:     query,
		TopK:      params.TopK,
	})
	if err != nil {
		return SearchProjectFileResult{Error: err.Error()}, nil
	}
	parents, err := readSearchParents(projectID, params.FileID, resp.GetParentIds())
	if err != nil {
		return SearchProjectFileResult{Error: err.Error()}, nil
	}
	return SearchProjectFileResult{Parents: parents}, nil
}

func NewSearchProjectFileTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"SearchProjectFile",
		"Search content inside an uploaded large file that is not directly shown in chat history. Use file_id from the file metadata. Returns matched parent blocks with parent_id and content.",
		SearchProjectFileFunc,
	)
}

func readSearchParents(projectID int64, fileID int64, parentIDs []int64) ([]SearchProjectFileParent, error) {
	staticRoot := filepath.Dir(filepath.Clean(conf.GetConf().ShareDir.ShareDir))
	parents := make([]SearchProjectFileParent, 0, len(parentIDs))
	for _, parentID := range parentIDs {
		if parentID <= 0 {
			continue
		}
		path := filepath.Join(
			staticRoot,
			"document",
			strconv.FormatInt(projectID, 10),
			strconv.FormatInt(fileID, 10),
			"parents",
			strconv.FormatInt(parentID, 10)+".txt",
		)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read parent block %d failed: %w", parentID, err)
		}
		parents = append(parents, SearchProjectFileParent{
			ParentID: parentID,
			Content:  string(content),
		})
	}
	return parents, nil
}
