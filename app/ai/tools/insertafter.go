package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type InsertAfterParams struct {
	Path    string `json:"path" jsonschema:"description=文件路径相对于项目根目录的相对路径，不能包含 .. 或使用绝对路径"`
	Anchor  string `json:"anchor" jsonschema:"description=锚点文本，必须在文件中唯一出现；内容会插入到该文本之后"`
	Content string `json:"content" jsonschema:"description=要插入的内容，换行和缩进需要自行包含"`
}

type InsertAfterResult struct {
	Ok    bool   `json:"ok" jsonschema:"description=是否成功插入"`
	Error string `json:"error" jsonschema:"description=错误信息"`
}

func InsertAfterFunc(ctx context.Context, params *InsertAfterParams) (InsertAfterResult, error) {
	store, err := projectStoreFromContext(ctx)
	if err != nil {
		return InsertAfterResult{Ok: false, Error: err.Error()}, nil
	}
	if err := store.InsertAfter(ctx, params.Path, params.Anchor, params.Content); err != nil {
		return InsertAfterResult{Ok: false, Error: err.Error()}, nil
	}
	return InsertAfterResult{Ok: true}, nil
}

func NewInsertAfterTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"InsertAfter",
		"在文件中的唯一锚点文本之后插入内容。用于追加 import、路由、样式块等小范围修改。",
		InsertAfterFunc)
}
