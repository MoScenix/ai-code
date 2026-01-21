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

type ViewFileParams struct {
	Path string `json:"path" jsonschema:"description=你要查看的文件所在目录相对于项目根目录的路径"`
	Name string `json:"name" jsonschema:"description=你要查看的文件名"`
}
type ViewFileResult struct {
	FileExist bool   `json:"file_exist" jsonschema:"description=文件是否存在"`
	Content   string `json:"content" jsonschema:"description=文件内容"`
	Error     string `json:"error" jsonschema:"description=错误信息"`
}

func ViewFileFunc(ctx context.Context, params *ViewFileParams) (ViewFileResult, error) {
	projectPath := filepath.Join(conf.GetConf().ShareDir.ShareDir, ctx.Value(lutils.ProjectRootPath).(string))
	target := filepath.Join(projectPath, params.Path)
	target = filepath.Join(target, params.Name)
	if !IsSubPathAbs(target, projectPath, false) {
		return ViewFileResult{
			Error: "路径不合法",
		}, nil
	}
	_, err := os.Stat(target)
	if err != nil {
		return ViewFileResult{
			FileExist: false,
			Error:     err.Error(),
		}, nil
	}
	content, err := os.ReadFile(target)
	if err != nil {
		return ViewFileResult{
			Error: err.Error(),
		}, err
	}
	return ViewFileResult{
		FileExist: true,
		Content:   string(content),
	}, nil
}

func NewViewFileTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"ViewFile",
		"查看一个项目文件的内容",
		ViewFileFunc)
}
