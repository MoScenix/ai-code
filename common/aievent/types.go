package aievent

import (
	"fmt"
	"strings"
	"time"
)

type ControlType string

const (
	ControlPush   ControlType = "push"
	ControlCancel ControlType = "cancel"
)

type EventType string

const (
	EventPush       EventType = "push"
	EventCancel     EventType = "cancel"
	EventAnswer     EventType = "answer"
	EventAccepted   EventType = "accepted"
	EventAgentStart EventType = "agent_start"
	EventMessage    EventType = "message"
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventQuestion   EventType = "question"
	EventDone       EventType = "done"
	EventCancelled  EventType = "cancelled"
	EventError      EventType = "error"
)

const (
	ProjectStatusQueued        = "queued"
	ProjectStatusRunning       = "running"
	ProjectStatusWaitingAnswer = "waiting_answer"
	ProjectStatusInterrupted   = "interrupted"
	ProjectStatusDone          = "done"
	ProjectStatusCancelled     = "cancelled"
	ProjectStatusError         = "error"
)

type ControlEvent struct {
	ProjectID string            `json:"project_id"`
	Type      ControlType       `json:"type"`
	Content   string            `json:"content,omitempty"`
	Reason    string            `json:"reason,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	CreatedAt int64             `json:"created_at"`
}

type TaskEvent struct {
	ProjectID string         `json:"project_id"`
	Type      EventType      `json:"type"`
	Agent     string         `json:"agent,omitempty"`
	Content   string         `json:"content,omitempty"`
	TargetID  string         `json:"target_id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Status    string         `json:"status,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt int64          `json:"created_at"`
}

type ProjectState struct {
	Status            string             `json:"status"`
	Agent             string             `json:"agent,omitempty"`
	LastEventID       string             `json:"last_event_id,omitempty"`
	CheckpointID      string             `json:"checkpoint_id,omitempty"`
	PendingInterrupts []PendingInterrupt `json:"pending_interrupts,omitempty"`
	Message           string             `json:"message,omitempty"`
	Buffer            string             `json:"buffer,omitempty"`
	IsCancelled       bool               `json:"is_cancelled,omitempty"`
	UpdatedAt         int64              `json:"updated_at"`
}

type PendingInterrupt struct {
	ID      string         `json:"id"`
	Agent   string         `json:"agent,omitempty"`
	Content string         `json:"content,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

func NewControl(projectID string, typ ControlType) ControlEvent {
	return ControlEvent{
		ProjectID: projectID,
		Type:      typ,
		CreatedAt: time.Now().UnixMilli(),
	}
}

func NewEvent(projectID string, typ EventType) TaskEvent {
	return TaskEvent{
		ProjectID: projectID,
		Type:      typ,
		CreatedAt: time.Now().UnixMilli(),
	}
}

func ControlKey(projectID string) string {
	return StreamKey(projectID)
}

func EventKey(projectID string) string {
	return StreamKey(projectID)
}

func StreamKey(projectID string) string {
	return projectKey(projectID, "stream")
}

func CursorKey(projectID string) string {
	return projectKey(projectID, "cursor")
}

func ActiveTaskKey(projectID string) string {
	return projectKey(projectID, "active_task")
}

func RunningStateKey(projectID string) string {
	return projectKey(projectID, "state")
}

func CheckpointKey(projectID string) string {
	return projectKey(projectID, "checkpoint")
}

func GraphCheckpointKey(projectID string) string {
	return projectKey(projectID, "graph_checkpoint")
}

func projectKey(projectID, suffix string) string {
	projectID = strings.Trim(projectID, ":")
	if projectID == "" {
		return fmt.Sprintf("project:unknown:%s", suffix)
	}
	return fmt.Sprintf("project:%s:%s", projectID, suffix)
}
