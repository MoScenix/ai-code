package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type EditFileParams struct {
	Path string `json:"path" jsonschema:"description=文件路径相对于项目根目录的相对路径，不能包含 .. 或使用绝对路径"`
	Old  string `json:"old" jsonschema:"description=需要替换的原始文本，必须在文件中唯一出现；如果不唯一，请提供更长上下文"`
	New  string `json:"new" jsonschema:"description=替换后的新文本"`
}

type EditFileResult struct {
	Ok    bool   `json:"ok" jsonschema:"description=是否成功编辑"`
	Error string `json:"error" jsonschema:"description=错误信息"`
}

func EditFileFunc(ctx context.Context, params *EditFileParams) (EditFileResult, error) {
	store, err := projectStoreFromContext(ctx)
	if err != nil {
		return EditFileResult{Ok: false, Error: err.Error()}, nil
	}
	if err := store.EditFile(ctx, params.Path, params.Old, params.New); err != nil {
		return EditFileResult{Ok: false, Error: err.Error()}, nil
	}
	return EditFileResult{Ok: true}, nil
}

func NewEditFileTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"EditFile",
		"精确替换文件中的一段文本。old 必须唯一匹配；用于修改已有文件，避免全量覆盖大文件。",
		EditFileFunc)
}
