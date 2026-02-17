package mcp

import (
	// "context"
	// "github.com/google/uuid"
	"encoding/json"
	"fmt"
	"time"
)

type Result struct {
	Text string
	Type string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func PrintTaskAsJSON(task *Task) {
	b, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling task:", err)
		return
	}
	fmt.Println(string(b))
}

func NewRunID() string {
	return fmt.Sprintf("run-%d", time.Now().UnixNano())
}

func NewTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

// based upon JSON-RPC
// ChatMessage represents incoming chat JSON
type ChatMessage struct {
	Text string `json:"text" binding:"required" example:"Analyze scenario, List all machinery, Categorize machinery"`
	Type string `json:"type" example:"user"` // user/tool/bot
}

type LLMMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type LLMParams struct {
    Messages []LLMMessage `json:"messages"`
}

type JsonRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
    ID      int         `json:"id"`
}

type JsonRPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    Result  json.RawMessage `json:"result"`
    Error   interface{}     `json:"error,omitempty"`
    ID      int             `json:"id"`
}

// --- Function to wrap text into JSON-RPC request ---

func MakeLLMJsonRPCRequest(userText string, id int) (*JsonRPCRequest, error) {
	messages := []LLMMessage{
		{
			Role:    "system",
			Content: "You are a risk assessment assistant.",
		},
		{
			Role:    "user",
			Content: userText,
		},
	}

	rpcReq := &JsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "llm/message",
		Params:  LLMParams{Messages: messages},
		ID:      id,
	}

	return rpcReq, nil
}

// --- Optional helper to print JSON string ---
func RpcRequestToJSON(rpcReq *JsonRPCRequest) (string, error) {
	data, err := json.MarshalIndent(rpcReq, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- SendToLLM with mock fallback ---

// --- SendToLLM with MCP types ---
func SendToLLM(apiURL string, rpcReq *JsonRPCRequest) (*JsonRPCResponse, error) {
    // 1️⃣ Mock fallback if apiURL is empty
    if apiURL == "" {
        // Prepare mock result
        mockResult := map[string]interface{}{
            "text":      "This is a mock LLM response for testing.",
            "tools_used": []string{},
        }

        resultBytes, err := json.Marshal(mockResult)
        if err != nil {
            return nil, err
        }

        return &JsonRPCResponse{
            JSONRPC: "2.0",
            Result:  resultBytes, // MCP expects []byte / json.RawMessage
            ID:      rpcReq.ID,
        }, nil
    }

    // 2️⃣ Real MCP / LLM call (example)
    resp, err := SendToLLM(apiURL, rpcReq)
    if err != nil {
        return nil, err
    }

    return resp, nil
}

func BuildRunWithOutput(input string) *Run {
	runID := NewRunID()
	fmt.Println("BuildRunWithOutput", input)
	output := &RunOutput{
		ID:        runID,
		Answer:    "",
		Machinery: []string{},
	}

	mainTask := &Task{
		ID:        NewTaskID(),
		RunID:     runID,
		Type:      "LLM",
		Input:     "Analyze scenario",
		Status:    "pending",
		CreatedAt: time.Now(),
		RunOutput: output,
	}

	listMachinery := &Task{
		ID:        NewTaskID(),
		RunID:     runID,
		Type:      "LLM",
		Input:     "List all machinery",
		Status:    "pending",
		CreatedAt: time.Now(),
		RunOutput: output,
	}

	categorizeMachinery := &Task{
		ID:        NewTaskID(),
		RunID:     runID,
		Type:      "Tool",
		Input:     "Categorize machinery",
		Status:    "pending",
		CreatedAt: time.Now(),
		RunOutput: output,
	}

	mainTask.Subtasks = []*Task{listMachinery, categorizeMachinery}

	PrintTaskAsJSON(mainTask)

	return &Run{
		ID:        runID,
		CreatedAt: time.Now(),
		Tasks:     []*Task{mainTask},
	}
}

func ExecuteTask(task *Task) {
	task.Status = "running"

	// LLM tasks
	if task.Type == "LLM" {
		fakeOutput := fmt.Sprintf("Fake LLM response for: %s", task.Input)
		task.Output = fakeOutput

		// Update shared output object
		switch task.Input {
		case "Analyze scenario":
			task.RunOutput.Answer = fakeOutput
		case "List all machinery":
			// Append machinery to the array
			task.RunOutput.Machinery = append(task.RunOutput.Machinery, "Excavator", "Crane", "Bulldozer")
		}
	}

	// Tool tasks
	if task.Type == "Tool" && task.Input == "Categorize machinery" {
		// Example: transform machinery array to categories
		// Here we just simulate by prefixing type
		categorized := []string{}
		for _, m := range task.RunOutput.Machinery {
			categorized = append(categorized, m+": Heavy Equipment")
		}
		task.RunOutput.Machinery = categorized
		task.Output = "Categorized machinery"
	}

	task.Status = "done"
	fmt.Println("Task done:", task.Input, "Output:", task.Output)

	for _, sub := range task.Subtasks {
		ExecuteTask(sub)
	}
}

// func HandleText(ctx context.Context, text string, msgType string) (*Result, error) {
//     log.Println("Received task:", text, "Type:", msgType)

//     task := Task{
//         ID:   NewRunID(),
//         Text: text,
//         Type: msgType,
//     }
//     log.Println("Created Task with ID:", task.ID)

//     output, err := runTask(ctx, task)
//     if err != nil {
//         log.Println("Error in runTask:", err)
//         return nil, err
//     }

//     log.Println("Task completed. Output:", output.Text)
//     return output, nil
// }

// func runTask(ctx context.Context, task Task) (*Result, error) {
//     log.Println("Running task:", task.ID)

//     if task.Type == "tool" {
//         log.Println("Executing tool step for task:", task.ID)
//         return &Result{
//             Text: "Tool executed: " + task.Text,
//             Type: "tool",
//         }, nil
//     }

//     log.Println("Building messages for LLM...")
//     messages := BuildLLMMessagesDynamic(task.Text)
// 	log.Println("messages:", messages)
//     // log messages for transparency
//     for i, m := range messages {
//         log.Printf("Message %d: role=%s, content=%s\n", i, m.Role, m.Content)
//     }

//     resp, err := callLLM(ctx, messages)
//     if err != nil {
//         log.Println("LLM call failed:", err)
//         return nil, err
//     }

//     log.Println("LLM responded:", resp)
//     return &Result{
//         Text: resp,
//         Type: "bot",
//     }, nil
// }
