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

type ReadDirParams struct {
	Path string `json:"path" jsonschema:"description=要查看的文件夹相对于项目根目录的相对路径，不能包含 ..或者使用绝对路径"`
}

type ReadDirResult struct {
	Error string     `json:"error" jsonschema:"description=错误信息"`
	Files []FileItem `json:"files" jsonschema:"description=文件夹下的文件列表"`
}
type FileItem struct {
	Name  string `json:"name" jsonschema:"description=文件名"`
	isDir bool   `json:"is_dir" jsonschema:"description=是否是文件夹"`
}

func ReadDirFunc(ctx context.Context, params *ReadDirParams) (ReadDirResult, error) {
	projectPath := filepath.Join(conf.GetConf().ShareDir.ShareDir, ctx.Value(lutils.ProjectRootPath).(string))
	target := filepath.Join(projectPath, params.Path)
	if !IsSubPathAbs(target, projectPath, true) {
		return ReadDirResult{
			Error: "路径不合法",
		}, nil
	}
	files, err := os.ReadDir(target)
	if err != nil {
		return ReadDirResult{
			Error: err.Error(),
		}, nil
	}
	result := ReadDirResult{}
	for _, file := range files {
		result.Files = append(result.Files, FileItem{
			Name:  file.Name(),
			isDir: file.IsDir(),
		})
	}
	return result, nil
}
func NewReadDirTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"ReadDir",
		"查看一个文件夹下的文件列表",
		ReadDirFunc)
}
