package mcp

import "time"

type LLMResponse struct {
	Text string
	Type string // "bot" or "tool"
}

type Run struct {
    ID        string
    CreatedAt time.Time
    Tasks     []*Task
}

type RunOutput struct {
    ID        string   `json:"id"`         // the RunID
    Answer    string   `json:"answer"`     // high-level response from LLM
    Machinery []string `json:"machinery"`  // will be populated by tool call
}

type Task struct {
    ID        string
    RunID     string
    Type      string // "LLM" or "Tool"
    Input     string
    Output    string
    Status    string // "pending", "running", "done"
    Subtasks  []*Task
    CreatedAt time.Time

    RunOutput *RunOutput // reference to the shared output object
}
