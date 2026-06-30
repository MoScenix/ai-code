package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type WriteFileParams struct {
	FilePath string `json:"file_path" jsonschema:"description=The path to the file to write"`
	Content  string `json:"content" jsonschema:"description=The file content to write"`
}

func WriteFileFunc(ctx context.Context, params *WriteFileParams) (string, error) {
	store, err := projectStoreFromContext(ctx)
	if err != nil {
		return "", err
	}
	if params == nil {
		return recoverableFilesystemMessage("write_file_missing_params", "missing params"), nil
	}
	if err := store.WriteFile(ctx, params.FilePath, []byte(params.Content)); err != nil {
		return recoverableFilesystemMessage("write_file_failed", err.Error()), nil
	}
	return successfulFilesystemMessage(fmt.Sprintf("Updated file %s", params.FilePath)), nil
}

func NewWriteFileTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"write_file",
		"Write a project-relative file. If writing fails, returns a JSON error result instead of failing the agent run.",
		WriteFileFunc,
	)
}
