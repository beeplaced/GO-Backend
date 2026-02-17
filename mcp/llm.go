package mcp

import (
	"context"
	"fmt"
)

func callLLM(ctx context.Context, messages []Message) (string, error) {
	// For now, just echo the user message
	userMessage := ""
	for _, m := range messages {
		if m.Role == "user" {
			userMessage = m.Content
		}
	}
	return "LLM simulated response to: " + userMessage, nil
}

func CallLLMStub(message Message) string {
	return fmt.Sprintf("Fake LLM output for input: %s", message.Content)
}
