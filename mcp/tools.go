package mcp

import (
	"strings"
)

type Tool struct {
    Name        string
    Description string
    Keywords    []string // used for matching user input
}

var tools = []Tool{
    {Name: "analyze_image", Keywords: []string{"analyze image", "image analysis"}},
    {Name: "categorize_machinery", Keywords: []string{"categorize machinery", "machinery categorization"}},
    {Name: "summarize_report", Keywords: []string{"summary", "summarize report"}},
}

func DetermineTools(userInput string) []string {
    userInputLower := strings.ToLower(userInput)
    selected := []string{}

    for _, tool := range tools {
        for _, kw := range tool.Keywords {
            if strings.Contains(userInputLower, kw) {
                selected = append(selected, tool.Name)
                break
            }
        }
    }

    return selected
}
