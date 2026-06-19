package service

import (
	"context"

	aitask "github.com/MoScenix/ai-code/app/ai/task"
	aiworkpool "github.com/MoScenix/ai-code/app/ai/workpool"
	"github.com/MoScenix/ai-code/common/rpcmeta"
	"github.com/cloudwego/kitex/pkg/klog"
)

type ChatService struct {
	ctx context.Context
}

// NewChatService new ChatService
func NewChatService(ctx context.Context) *ChatService {
	return &ChatService{ctx: ctx}
}

func (s *ChatService) Run(projectID string) (bool, error) {
	runCtx := context.WithoutCancel(s.ctx)
	identity := rpcmeta.FromContext(runCtx)
	task := aitask.NewChatTask(projectID, aitask.WithIdentity(identity))
	if err := task.Enqueue(runCtx); err != nil {
		klog.CtxErrorf(runCtx, "enqueue ai task failed: project_id=%s err=%v", projectID, err)
		return false, err
	}

	p, err := aiworkpool.Get()
	if err != nil {
		klog.CtxErrorf(runCtx, "get ai workpool failed: project_id=%s err=%v", projectID, err)
		return false, err
	}
	if err := p.Submit(runCtx, task); err != nil {
		klog.CtxErrorf(runCtx, "submit ai task failed: project_id=%s err=%v", projectID, err)
		return false, err
	}
	klog.CtxInfof(runCtx, "ai task queued: project_id=%s", projectID)
	return true, nil
}
