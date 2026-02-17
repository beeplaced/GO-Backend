package database

import (
    "encoding/json"
    "os"
    "fmt"
    "time"
)

const dbFile = "./database/data.json"

func NewID() string {
    return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

type Item struct {
    ID   string                 `json:"id"`
    Data map[string]interface{} `json:"data"`
}

// Read existing array from file
func ReadDB() ([]Item, error) {
    if _, err := os.Stat(dbFile); os.IsNotExist(err) {
        return []Item{}, nil
    }

    content, err := os.ReadFile(dbFile)
    if err != nil {
        return nil, err
    }

    var items []Item
    err = json.Unmarshal(content, &items)
    if err != nil {
        return nil, err
    }

    return items, nil
}

// Write array back to file
func WriteDB(items []Item) error {
    content, err := json.MarshalIndent(items, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(dbFile, content, 0644)
}

func GetByID(id string) (*Item, bool, error) {
    items, err := ReadDB()
    if err != nil {
        return nil, false, err
    }

    for _, item := range items {
        if item.ID == id {
            return &item, true, nil
        }
    }

    return nil, false, nil
}