package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/eino/schema"
)

const defaultHistoryLimit int64 = 20

func LoadChatHistory(ctx context.Context, appID int64) ([]*schema.Message, error) {
	return LoadChatHistoryWithLimit(ctx, appID, defaultHistoryLimit)
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
