package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MoScenix/ai-code/app/ai/conf"
	lutils "github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type MkPathParams struct {
	Path string `json:"path" jsonschema:"description=要创建的目录或文件相对于项目根目录的相对路径，不要包含 ..或者使用绝对路径，文件夹以"/"结尾`
}
type MkPathResult struct {
	Ok    bool   `json:"ok" jsonschema:"description=目录是否成功创建或已存在"`
	Error string `json:"error" jsonschema:"description=错误信息"`
}

func MkdirFunc(ctx context.Context, params *MkPathParams) (MkPathResult, error) {
	projectPath := filepath.Join(conf.GetConf().ShareDir.ShareDir, ctx.Value(lutils.ProjectRootPath).(string))

	target := filepath.Join(projectPath, params.Path)
	fmt.Println("Mkpath" + target)
	isDir := strings.HasSuffix(params.Path, "/")
	if !IsSubPathAbs(target, projectPath, false) {
		return MkPathResult{
			Ok:    false,
			Error: "路径不合法,项目路径是" + projectPath + "你的路径是" + target,
		}, nil
	}
	if isDir {
		err := os.MkdirAll(target, os.ModePerm)
		if err != nil {
			return MkPathResult{
				Ok:    false,
				Error: err.Error(),
			}, nil
		}
		return MkPathResult{
			Ok:    true,
			Error: "",
		}, nil
	} else {
		err := os.MkdirAll(filepath.Dir(target), os.ModePerm)
		if err != nil {
			return MkPathResult{
				Ok:    false,
				Error: err.Error(),
			}, nil
		}
		file, err := os.Create(target)
		if err != nil {
			return MkPathResult{
				Ok:    false,
				Error: err.Error(),
			}, nil
		}
		defer file.Close()
		return MkPathResult{
			Ok:    true,
			Error: "",
		}, nil
	}
}

func NewMkPath() (tool.InvokableTool, error) {
	return utils.InferTool(
		"MkPath",
		`在项目根目录下创建一个文件夹或文件（自动创建缺失的父目录）,如果创建的是目录以"/"结尾`,
		MkdirFunc)
}
