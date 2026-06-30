package tools

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type EditFileParams struct {
	FilePath   string `json:"file_path" jsonschema:"description=The path to the file to modify"`
	OldString  string `json:"old_string" jsonschema:"description=The text to replace"`
	NewString  string `json:"new_string" jsonschema:"description=The text to replace it with"`
	ReplaceAll bool   `json:"replace_all" jsonschema:"description=Replace all occurrences of old_string (default false),default=false"`
}

func EditFileFunc(ctx context.Context, params *EditFileParams) (string, error) {
	store, err := projectStoreFromContext(ctx)
	if err != nil {
		return "", err
	}
	if params == nil {
		return recoverableFilesystemMessage("edit_file_missing_params", "missing params"), nil
	}
	if params.OldString == "" {
		return recoverableFilesystemMessage("empty_old_text", "old_string 不能为空。请先读取文件内容，确认要替换的完整文本。"), nil
	}

	data, err := store.ReadFile(ctx, params.FilePath)
	if err != nil {
		return recoverableFilesystemMessage("edit_file_read_failed", err.Error()), nil
	}

	content := string(data)
	count := strings.Count(content, params.OldString)
	if count == 0 {
		return recoverableFilesystemMessage("text_not_found", "目标文本未找到。请先读取文件内容，确认 old_string 完全一致后再调用 edit_file。"), nil
	}
	if count > 1 && !params.ReplaceAll {
		return recoverableFilesystemMessage("text_not_unique", "目标文本出现多次。请提供更长上下文让 old_string 唯一，或设置 replace_all=true。"), nil
	}

	next := strings.Replace(content, params.OldString, params.NewString, 1)
	if params.ReplaceAll {
		next = strings.ReplaceAll(content, params.OldString, params.NewString)
	}
	if err := store.WriteFile(ctx, params.FilePath, []byte(next)); err != nil {
		return recoverableFilesystemMessage("edit_file_write_failed", err.Error()), nil
	}

	replaced := 1
	if params.ReplaceAll {
		replaced = count
	}
	if replaced > 1 {
		return successfulFilesystemMessage("Successfully replaced all matching strings."), nil
	}
	return successfulFilesystemMessage("Successfully replaced the string."), nil
}

func NewEditFileTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"edit_file",
		"精确替换文件中的一段文本。old_string 必须完全匹配文件内容；如果失败会返回 JSON 错误结果，请先 read_file 确认内容后再重试。",
		EditFileFunc)
}
