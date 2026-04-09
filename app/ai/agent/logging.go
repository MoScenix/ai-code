package agent

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func LogStreamChunk(scope string, msg *schema.Message) {
	if msg == nil {
		fmt.Printf("[ai-agent] %s recv nil message chunk\n", scope)
		return
	}

	fmt.Printf("[ai-agent] %s chunk role=%s content=%s tool_calls=%d tool_call_id=%s tool_name=%s\n",
		scope,
		msg.Role,
		CompactAgentLog(msg.Content),
		len(msg.ToolCalls),
		CompactAgentLog(msg.ToolCallID),
		CompactAgentLog(msg.ToolName),
	)

	for idx, tc := range msg.ToolCalls {
		fmt.Printf("[ai-agent] %s tool_call[%d] id=%s type=%s name=%s args=%s\n",
			scope,
			idx,
			CompactAgentLog(tc.ID),
			CompactAgentLog(tc.Type),
			CompactAgentLog(tc.Function.Name),
			CompactAgentLog(tc.Function.Arguments),
		)
	}
}

func CompactAgentLog(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	if len(s) > 600 {
		return s[:600] + "...(truncated)"
	}
	if s == "" {
		return "<empty>"
	}
	return s
}
