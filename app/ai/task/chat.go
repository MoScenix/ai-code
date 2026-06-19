package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/graph"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/filestore/project"
	"github.com/MoScenix/ai-code/common/rpcmeta"
)

type ChatTask struct {
	ProjectID string

	ctx           context.Context
	runtime       *utils.RuntimeState
	identity      rpcmeta.Identity
	needResume    bool
	previousState aievent.ProjectState
}

type ChatTaskOption func(*ChatTask)

func WithIdentity(identity rpcmeta.Identity) ChatTaskOption {
	return func(t *ChatTask) {
		t.identity = identity
	}
}

func NewChatTask(projectID string, opts ...ChatTaskOption) *ChatTask {
	task := &ChatTask{
		ProjectID: strings.TrimSpace(projectID),
	}
	for _, opt := range opts {
		opt(task)
	}
	return task
}

func (t *ChatTask) Init(ctx context.Context) (context.Context, error) {
	if t.ctx != nil {
		return t.ctx, nil
	}

	ctx = rpcmeta.WithIdentity(ctx, t.identity)
	runCtx, cancel := context.WithCancel(ctx)
	runtime := utils.NewRuntimeState(cancel)

	store, err := project.NewDefaultStore(t.ProjectID)
	if err != nil {
		cancel()
		return nil, err
	}

	runCtx = utils.WithRuntimeState(runCtx, runtime)
	runCtx = utils.WithStringBuffer(runCtx, runtime.Buffer)
	runCtx = utils.WithCancelFunc(runCtx, cancel)
	runCtx = utils.WithProjectStore(runCtx, store)

	t.ctx = runCtx
	t.runtime = runtime
	if err := t.loadPlan(runCtx); err != nil {
		cancel()
		return nil, err
	}
	return runCtx, nil
}

func (t *ChatTask) Enqueue(ctx context.Context) error {
	runCtx, err := t.Init(ctx)
	if err != nil {
		return err
	}
	return t.markState(runCtx, aievent.ProjectStatusQueued)
}

func (t *ChatTask) Run(ctx context.Context) error {
	runCtx, err := t.Init(ctx)
	if err != nil {
		return err
	}
	defer t.runtime.Stop()

	if err := t.markState(runCtx, aievent.ProjectStatusRunning); err != nil {
		return err
	}

	if t.needResume {
		err = graph.Resume(runCtx)
	} else {
		err = graph.Run(runCtx)
	}
	if errors.Is(err, graph.ErrInterrupted) {
		return nil
	}
	if t.runtime.IsCancelled() && errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (t *ChatTask) Resume(ctx context.Context) error {
	return t.Run(ctx)
}

func (t *ChatTask) loadPlan(ctx context.Context) error {
	stateStore, ok := utils.StateStoreFromContext(ctx)
	if !ok || stateStore == nil {
		return nil
	}

	var state aievent.ProjectState
	ok, err := stateStore.Get(ctx, aievent.RunningStateKey(t.ProjectID), &state)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	t.previousState = state
	t.needResume = state.Status == aievent.ProjectStatusInterrupted && state.CheckpointID != "" && len(state.PendingInterrupts) > 0
	if buffer, ok := utils.StringBufferFromContext(ctx); ok && state.Buffer != "" {
		buffer.SetString(state.Buffer)
	}
	if cursor := pendingControlCursor(state); cursor != "" {
		utils.SetControlCursor(ctx, cursor)
	}
	return nil
}

func pendingControlCursor(state aievent.ProjectState) string {
	if len(state.PendingInterrupts) == 0 || state.PendingInterrupts[0].Payload == nil {
		return ""
	}
	return aievent.ControlCursor(state.PendingInterrupts[0].Payload)
}

func (t *ChatTask) markState(ctx context.Context, status string) error {
	stateStore, ok := utils.StateStoreFromContext(ctx)
	if !ok || stateStore == nil {
		return nil
	}
	if t.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}

	state := aievent.ProjectState{
		Status:            status,
		Agent:             t.previousState.Agent,
		LastEventID:       t.previousState.LastEventID,
		CheckpointID:      t.previousState.CheckpointID,
		PendingInterrupts: t.previousState.PendingInterrupts,
		Message:           t.previousState.Message,
		Buffer:            t.previousState.Buffer,
		IsCancelled:       false,
		UpdatedAt:         time.Now().UnixMilli(),
	}
	if state.Agent == "" && t.needResume {
		state.Agent = "Graph"
	}
	return stateStore.Set(ctx, aievent.RunningStateKey(t.ProjectID), state)
}
