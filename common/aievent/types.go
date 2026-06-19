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

const MaxEventContentRunes = 1200

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

const (
	PayloadInfo           = "info"
	PayloadADKInterruptID = "adk_interrupt_id"
	PayloadControlCursor  = "control_cursor"
	PayloadDesignerLastID = "designer_last_id"
	PayloadLastEventID    = "last_event_id"
)

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
	return projectKey(projectID, "control")
}

func EventKey(projectID string) string {
	return projectKey(projectID, "stream")
}

func StreamKey(projectID string) string {
	return EventKey(projectID)
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

func PayloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, ok := payload[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func NestedPayload(payload map[string]any, key string) map[string]any {
	if payload == nil {
		return nil
	}
	nested, ok := payload[key].(map[string]any)
	if !ok {
		return nil
	}
	return nested
}

func ADKInterruptID(payload map[string]any) string {
	return PayloadString(payload, PayloadADKInterruptID)
}

func ControlCursor(payload map[string]any) string {
	return PayloadString(payload, PayloadControlCursor)
}

func DesignerLastID(payload map[string]any) string {
	if value := PayloadString(payload, PayloadDesignerLastID); value != "" {
		return value
	}
	return PayloadString(payload, PayloadLastEventID)
}

func PendingInterruptTargetIDs(interrupt PendingInterrupt) map[string]bool {
	ids := make(map[string]bool)
	if id := strings.TrimSpace(interrupt.ID); id != "" {
		ids[id] = true
	}
	if id := ADKInterruptID(interrupt.Payload); id != "" {
		ids[id] = true
	}
	if id := ADKInterruptID(NestedPayload(interrupt.Payload, PayloadInfo)); id != "" {
		ids[id] = true
	}
	return ids
}

func PendingInterruptMatches(interrupt PendingInterrupt, targetID string) bool {
	targetID = strings.TrimSpace(targetID)
	return targetID != "" && PendingInterruptTargetIDs(interrupt)[targetID]
}

func PendingInterruptsMatch(interrupts []PendingInterrupt, targetID string) bool {
	for _, interrupt := range interrupts {
		if PendingInterruptMatches(interrupt, targetID) {
			return true
		}
	}
	return false
}

func TrimEventContent(content string) string {
	runes := []rune(content)
	if len(runes) <= MaxEventContentRunes {
		return content
	}
	return string(runes[:MaxEventContentRunes]) + fmt.Sprintf("\n... truncated %d chars", len(runes)-MaxEventContentRunes)
}
