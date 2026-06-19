package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/common/rpcmeta"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

const defaultHistoryLimit int64 = 20

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

	client, err := AppClient()
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

	client, err := AppClient()
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
		switch msg.Role {
		case "user":
			messages = append(messages, schema.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, schema.AssistantMessage(msg.Content, nil))
		case "system":
			messages = append(messages, schema.SystemMessage(msg.Content))
		default:
			messages = append(messages, schema.UserMessage(msg.Content))
		}
	}
	return messages, nil
}
