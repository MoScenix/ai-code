package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MoScenix/ai-code/app/ai/conf"
	lutils "github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type DeleteFile struct {
	Path string `json:"path" jsonschema:"description=需要删除的文件或文件夹路径相对于项目根目录的相对路径，不能包含 ..或者使用绝对路径"`
}
type DeleteFileResult struct {
	Ok    bool   `json:"success" jsonschema:"description=是否成功/本身就不存在"`
	Error string `json:"error" jsonschema:"description=错误信息"`
}

func DeleteFileFunc(ctx context.Context, params *DeleteFile) (DeleteFileResult, error) {
	projectPath := filepath.Join(conf.GetConf().ShareDir.ShareDir, ctx.Value(lutils.ProjectRootPath).(string))
	target := filepath.Join(projectPath, params.Path)
	fmt.Println("DeleteFile" + target)
	if !IsSubPathAbs(target, projectPath, false) {
		return DeleteFileResult{
			Ok:    false,
			Error: "路径不合法",
		}, nil
	}
	state, err := os.Stat(target)
	if err == nil {
		if state.IsDir() {
			err = os.RemoveAll(target)
		} else {
			err = os.Remove(target)
		}
		if err != nil {
			return DeleteFileResult{
				Ok:    false,
				Error: err.Error(),
			}, nil
		}
		return DeleteFileResult{
			Ok:    true,
			Error: "",
		}, nil
	}
	return DeleteFileResult{
		Ok:    true,
		Error: "",
	}, nil
}
func NewDeleteFileTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"DeleteFile",
		"删除一个文件或文件夹",
		DeleteFileFunc)
}
