package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/MoScenix/ai-code/app/ai/utils"
)

func InjectHistory(ctx context.Context, projectID string) (context.Context, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("project id is required")
	}

	messages, err := utils.LoadProjectChatHistory(ctx, projectID)
	if err != nil {
		return nil, err
	}
	ctx = utils.WithProjectID(ctx, projectID)
	ctx = utils.WithHistoryMessages(ctx, messages)
	return ctx, nil
}
