package mcp

import (
    "fmt"
)

func BuildFirstLLMMessage(task *Task) Message {
    systemPrompt := "You are a helpful risk assessment LLM. Analyze the input carefully."
    userPrompt := task.Input

    return Message{
        Role:    "system",
        Content: fmt.Sprintf("%s\nUser input: %s", systemPrompt, userPrompt),
    }
}

func BuildLLMMessagesDynamic(userText string) []Message {
    inputType := DetectInputType(userText)
    systemPrompt := BuildSystemPrompt(inputType)

    messages := []Message{
        {
            Role:    "system",
            Content: systemPrompt,
        },
        {
            Role:    "user",
            Content: userText,
        },
    }
    return messages
}

func DetectInputType(text string) string {
    // Very basic rules for MVP
    if len(text) == 0 {
        return "unknown"
    }
    lastChar := text[len(text)-1]
    switch lastChar {
    case '?':
        return "question"
    default:
        // Could use keywords like "please", "analyze", etc.
        if len(text) > 50 {
            return "observation"
        }
        return "request"
    }
}

func BuildSystemPrompt(inputType string) string {
    switch inputType {
    case "question":
        return "You are a helpful risk assessment LLM. Answer questions clearly and concisely."
    case "request":
        return "You are a proactive risk assessment LLM. Take action-oriented steps to address the request."
    case "observation":
        return "You are a risk assessment LLM. Analyze the information and summarize any potential risks."
    default:
        return "You are a helpful risk assessment LLM. Respond appropriately."
    }
}