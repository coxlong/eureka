package model

import (
	"encoding/json"
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	Parent    string    `json:"parent"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationMeta struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Model         string    `json:"model"`
	MaxTokens     int       `json:"max_tokens"`
	Temperature   float32   `json:"temperature"`
	CurrentNodeID string    `json:"current_node_id"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

func (c *ConversationMeta) MarshalJSON() ([]byte, error) {
	type Alias ConversationMeta
	return json.Marshal(struct {
		*Alias
		CreatedAt int64 `json:"created_at"`
		UpdatedAt int64 `json:"updated_at"`
	}{
		Alias:     (*Alias)(c),
		CreatedAt: c.CreatedAt.UnixMilli(),
		UpdatedAt: c.UpdatedAt.UnixMilli(),
	})
}
