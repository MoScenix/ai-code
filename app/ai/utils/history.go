package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/conf"
	"github.com/MoScenix/ai-code/app/ai/infra"
	"github.com/MoScenix/ai-code/common/rpcmeta"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

const defaultHistoryLimit int64 = 20

type fileMessageContent struct {
	FileID       int64  `json:"fileId"`
	Filename     string `json:"filename"`
	ContentType  string `json:"contentType"`
	Size         int64  `json:"size"`
	TextFilename string `json:"textFilename"`
	TextSize     int64  `json:"textSize"`
	IsBig        bool   `json:"isBig"`
	ChunkCount   int64  `json:"chunkCount,omitempty"`
	ParentCount  int64  `json:"parentCount,omitempty"`
	Text         string `json:"text,omitempty"`
}

func LoadChatHistory(ctx context.Context, appID int64) ([]*schema.Message, error) {
	return LoadChatHistoryWithLimit(ctx, appID, defaultHistoryLimit)
}

func AddAssistantMessage(ctx context.Context, appID int64, content string) error {
	return addChatMessage(ctx, appID, "assistant", content)
}

func AddUserMessage(ctx context.Context, appID int64, content string) error {
	return addChatMessage(ctx, appID, "user", content)
}

func addChatMessage(ctx context.Context, appID int64, role string, content string) error {
	content = strings.TrimSpace(content)
	userID, _ := rpcmeta.OperatorIDFromContext(ctx)
	if appID <= 0 || content == "" {
		return nil
	}

	client, err := infra.AppClient()
	if err != nil {
		klog.CtxErrorf(ctx, "get app client failed while saving ai message: app_id=%d role=%s err=%v", appID, role, err)
		return err
	}

	_, err = client.AddMessage(ctx, &rpcapp.AddMessageReq{
		AppId:   appID,
		UserId:  userID,
		Role:    role,
		Content: content,
	})
	if err != nil {
		klog.CtxErrorf(ctx, "save ai message failed: app_id=%d user_id=%d role=%s err=%v", appID, userID, role, err)
		return err
	}
	return err
}

func AddProjectAssistantMessage(ctx context.Context, projectID string, content string) error {
	appID, err := strconv.ParseInt(projectID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse project id %q: %w", projectID, err)
	}
	return AddAssistantMessage(ctx, appID, content)
}

func AddProjectUserMessage(ctx context.Context, projectID string, content string) error {
	appID, err := strconv.ParseInt(projectID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse project id %q: %w", projectID, err)
	}
	return AddUserMessage(ctx, appID, content)
}

func LoadProjectChatHistory(ctx context.Context, projectID string) ([]*schema.Message, error) {
	appID, err := strconv.ParseInt(projectID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse project id %q: %w", projectID, err)
	}
	return LoadChatHistory(ctx, appID)
}

func LoadChatHistoryWithLimit(ctx context.Context, appID int64, limit int64) ([]*schema.Message, error) {
	if limit <= 0 {
		limit = defaultHistoryLimit
	}

	client, err := infra.AppClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.ListAppMessage(ctx, &rpcapp.ListAppMessageReq{
		AppId:          appID,
		PageSize:       limit,
		LastCreateTime: time.Now().Add(20 * time.Second).Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return nil, err
	}

	messages := make([]*schema.Message, 0, len(resp.MessageList))
	for i := len(resp.MessageList) - 1; i >= 0; i-- {
		msg := resp.MessageList[i]
		content := msg.Content
		if msg.GetIsFile() {
			content = formatFileHistoryMessage(appID, msg.Content)
		}
		switch msg.Role {
		case "user":
			messages = append(messages, schema.UserMessage(content))
		case "assistant":
			messages = append(messages, schema.AssistantMessage(content, nil))
		case "system":
			messages = append(messages, schema.SystemMessage(content))
		default:
			messages = append(messages, schema.UserMessage(content))
		}
	}
	return messages, nil
}

func formatFileHistoryMessage(appID int64, raw string) string {
	var meta fileMessageContent
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return "用户上传了一个文件，但文件元信息解析失败。"
	}

	filename := strings.TrimSpace(meta.Filename)
	if filename == "" {
		filename = "未命名文件"
	}

	var builder strings.Builder
	if meta.IsBig {
		builder.WriteString("用户上传了一个大文件，这个文件内容不会直接显示在聊天记录中。\n")
		builder.WriteString("如需查看文件内容，请调用 SearchProjectFile 工具搜索这个文件。\n")
		builder.WriteString("文件元信息：\n")
		builder.WriteString(fmt.Sprintf("- file_id: %d\n", meta.FileID))
		builder.WriteString(fmt.Sprintf("- filename: %s\n", filename))
		if meta.ContentType != "" {
			builder.WriteString(fmt.Sprintf("- content_type: %s\n", meta.ContentType))
		}
		if meta.Size > 0 {
			builder.WriteString(fmt.Sprintf("- size_bytes: %d\n", meta.Size))
		}
		if meta.TextSize > 0 {
			builder.WriteString(fmt.Sprintf("- text_size_bytes: %d\n", meta.TextSize))
		}
		if meta.ParentCount > 0 {
			builder.WriteString(fmt.Sprintf("- parent_count: %d\n", meta.ParentCount))
		}
		if meta.ChunkCount > 0 {
			builder.WriteString(fmt.Sprintf("- chunk_count: %d\n", meta.ChunkCount))
		}
		return strings.TrimSpace(builder.String())
	}

	builder.WriteString("用户上传了一个文件，以下是文件内容。\n")
	builder.WriteString("文件元信息：\n")
	builder.WriteString(fmt.Sprintf("- file_id: %d\n", meta.FileID))
	builder.WriteString(fmt.Sprintf("- filename: %s\n", filename))
	if meta.ContentType != "" {
		builder.WriteString(fmt.Sprintf("- content_type: %s\n", meta.ContentType))
	}
	if meta.TextSize > 0 {
		builder.WriteString(fmt.Sprintf("- text_size_bytes: %d\n", meta.TextSize))
	}
	builder.WriteString("\n文件内容：\n")
	text := readSmallFileText(appID, meta)
	if strings.TrimSpace(text) == "" {
		builder.WriteString("（文件内容为空或未解析出文本）")
	} else {
		builder.WriteString(text)
	}
	return strings.TrimSpace(builder.String())
}

func readSmallFileText(appID int64, meta fileMessageContent) string {
	if meta.FileID <= 0 {
		return meta.Text
	}

	filename := strings.TrimSpace(meta.TextFilename)
	if filename == "" {
		filename = strings.TrimSpace(meta.Filename)
	}
	if filename == "" {
		return meta.Text
	}
	if strings.EqualFold(filepath.Ext(filename), ".pdf") {
		filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".txt"
	}

	staticRoot := filepath.Dir(filepath.Clean(conf.GetConf().ShareDir.ShareDir))
	path := filepath.Join(staticRoot, "document", strconv.FormatInt(appID, 10), strconv.FormatInt(meta.FileID, 10), filename)
	raw, err := os.ReadFile(path)
	if err != nil {
		return meta.Text
	}
	return strings.TrimSpace(string(raw))
}
