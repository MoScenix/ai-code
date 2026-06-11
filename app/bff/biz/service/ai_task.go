package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/bff/biz/dal/redis"
	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	rpcai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

func submitAITask(ctx context.Context, appID int64, message string) (bool, error) {
	if appID <= 0 {
		return false, fmt.Errorf("appId is required")
	}
	if strings.TrimSpace(message) != "" {
		if err := addUserMessage(ctx, appID, message); err != nil {
			return false, err
		}
	}
	return submitAI(ctx, appID)
}

func pushAIEvent(ctx context.Context, appID int64, content string) (string, error) {
	if appID <= 0 {
		return "", fmt.Errorf("appId is required")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("content is required")
	}
	return addTaskEvent(ctx, appID, aievent.TaskEvent{
		ProjectID: projectID(appID),
		Type:      aievent.EventPush,
		Content:   content,
	})
}

func answerAIQuestion(ctx context.Context, appID int64, content string, targetID string) (bool, error) {
	if appID <= 0 {
		return false, fmt.Errorf("appId is required")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return false, fmt.Errorf("content is required")
	}

	state, ok, err := loadAIState(ctx, appID)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, fmt.Errorf("ai task state not found")
	}
	targetID = strings.TrimSpace(targetID)
	if targetID == "" && len(state.PendingInterrupts) > 0 {
		targetID = state.PendingInterrupts[0].ID
	}
	if targetID == "" {
		return false, fmt.Errorf("pending interrupt target not found")
	}

	if _, err := addTaskEvent(ctx, appID, aievent.TaskEvent{
		ProjectID: projectID(appID),
		Type:      aievent.EventAnswer,
		Content:   content,
		TargetID:  targetID,
	}); err != nil {
		return false, err
	}

	if state.Status == aievent.ProjectStatusInterrupted {
		return submitAI(ctx, appID)
	}
	return false, nil
}

func cancelAIEvent(ctx context.Context, appID int64, reason string) (string, error) {
	if appID <= 0 {
		return "", fmt.Errorf("appId is required")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "cancelled"
	}
	return addTaskEvent(ctx, appID, aievent.TaskEvent{
		ProjectID: projectID(appID),
		Type:      aievent.EventCancel,
		Content:   reason,
	})
}

func loadAIState(ctx context.Context, appID int64) (aievent.ProjectState, bool, error) {
	stateStore, err := stateStore()
	if err != nil {
		return aievent.ProjectState{}, false, err
	}
	var state aievent.ProjectState
	ok, err := stateStore.Get(ctx, aievent.RunningStateKey(projectID(appID)), &state)
	return state, ok, err
}

func listAIEvents(ctx context.Context, appID int64, lastID string, blockMS int64, count int64) (*lapp.AIEvents, error) {
	if appID <= 0 {
		return nil, fmt.Errorf("appId is required")
	}
	lastID = strings.TrimSpace(lastID)
	if lastID == "" {
		lastID = "0"
	}
	if count <= 0 {
		count = 50
	}
	block := time.Duration(blockMS) * time.Millisecond
	if block < 0 {
		block = 0
	}

	store, err := streamStore()
	if err != nil {
		return nil, err
	}
	messages, err := store.Read(ctx, aievent.StreamKey(projectID(appID)), lastID, redisstream.ReadOptions{
		Block: block,
		Count: int64(count),
	})
	if err != nil {
		return nil, err
	}

	resp := &lapp.AIEvents{Events: make([]*lapp.AIEvent, 0, len(messages)), LastId: lastID}
	for _, msg := range messages {
		event, err := redisstream.Decode[aievent.TaskEvent](msg)
		if err != nil {
			continue
		}
		resp.Events = append(resp.Events, toAIEvent(msg.ID, event))
		resp.LastId = msg.ID
	}
	return resp, nil
}

func addUserMessage(ctx context.Context, appID int64, content string) error {
	userID, _ := contextUserID(ctx)
	_, err := rpc.AppClient.AddMessage(ctx, &rpcapp.AddMessageReq{
		AppId:   appID,
		UserId:  userID,
		Content: content,
		Role:    "user",
	})
	return err
}

func submitAI(ctx context.Context, appID int64) (bool, error) {
	stream, err := rpc.AiClient.Chat(ctx, &rpcai.AiReq{ProjectId: projectID(appID)})
	if err != nil {
		return false, err
	}
	defer stream.Close()

	accepted := false
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}
		accepted = resp.GetAnswer() == "true"
	}
	return accepted, nil
}

func addTaskEvent(ctx context.Context, appID int64, event aievent.TaskEvent) (string, error) {
	store, err := streamStore()
	if err != nil {
		return "", err
	}
	if event.ProjectID == "" {
		event.ProjectID = projectID(appID)
	}
	if event.CreatedAt == 0 {
		event.CreatedAt = time.Now().UnixMilli()
	}
	return store.Add(ctx, aievent.StreamKey(projectID(appID)), event)
}

func streamStore() (redisstream.Store, error) {
	return redisstream.NewRedisStore(redis.RedisClient, "ai")
}

func stateStore() (*redisstate.Store, error) {
	return redisstate.NewStore(redis.RedisClient, "ai")
}

func projectID(appID int64) string {
	return strconv.FormatInt(appID, 10)
}

func toAIState(exists bool, state aievent.ProjectState) *lapp.AIState {
	resp := &lapp.AIState{
		Exists:       exists,
		Status:       state.Status,
		Agent:        state.Agent,
		LastEventId:  state.LastEventID,
		CheckpointId: state.CheckpointID,
		Message:      state.Message,
		Buffer:       state.Buffer,
		IsCancelled:  state.IsCancelled,
		UpdatedAt:    state.UpdatedAt,
	}
	resp.PendingInterrupts = make([]*lapp.AIPendingInterrupt, 0, len(state.PendingInterrupts))
	for _, interrupt := range state.PendingInterrupts {
		resp.PendingInterrupts = append(resp.PendingInterrupts, &lapp.AIPendingInterrupt{
			Id:          interrupt.ID,
			Agent:       interrupt.Agent,
			Content:     interrupt.Content,
			PayloadJson: marshalPayload(interrupt.Payload),
		})
	}
	return resp
}

func toAIEvent(id string, event aievent.TaskEvent) *lapp.AIEvent {
	return &lapp.AIEvent{
		Id:          id,
		ProjectId:   event.ProjectID,
		Type:        string(event.Type),
		Agent:       event.Agent,
		Content:     event.Content,
		TargetId:    event.TargetID,
		Name:        event.Name,
		Status:      event.Status,
		PayloadJson: marshalPayload(event.Payload),
		CreatedAt:   event.CreatedAt,
	}
}

func marshalPayload(payload map[string]any) string {
	if len(payload) == 0 {
		return ""
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func contextUserID(ctx context.Context) (int64, bool) {
	switch v := ctx.Value(utils.UserIdKey).(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		return id, err == nil
	default:
		return 0, false
	}
}
