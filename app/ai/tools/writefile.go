package tools

import (
	"context"
	"os"
	"path/filepath"

	"github.com/MoScenix/ai-code/app/ai/conf"
	lutils "github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type WriteFileParams struct {
	Path    string `json:"path" jsonschema:"description=文件路径相对于项目根目录的相对路径，不能包含 ..或者使用绝对路径"`
	Content string `json:"content" jsonschema:"description=要写入的文件内容"`
}
type WriteFileResult struct {
	Error string `json:"error" jsonschema:"description=错误信息"`
	Ok    bool   `json:"ok" jsonschema:"description=文件是否成功写入"`
}

func WriteFileFunc(ctx context.Context, params *WriteFileParams) (WriteFileResult, error) {
	projectPath := filepath.Join(conf.GetConf().ShareDir.ShareDir, ctx.Value(lutils.ProjectRootPath).(string))
	target := filepath.Join(projectPath, params.Path)
	if !IsSubPathAbs(target, projectPath, false) {
		return WriteFileResult{
			Ok:    false,
			Error: "路径不合法",
		}, nil
	}
	err := os.WriteFile(target, []byte(params.Content), 0644)
	if err != nil {
		return WriteFileResult{
			Ok:    false,
			Error: err.Error(),
		}, nil
	}
	return WriteFileResult{
		Ok: true,
	}, nil
}
func NewWriteFileTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"WriteFile",
		"将内容写入一个文件,完全覆盖，文件不存在会创建(不包含父级目录)",
		WriteFileFunc)
}
