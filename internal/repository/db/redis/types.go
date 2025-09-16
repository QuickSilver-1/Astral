package redis

import "time"

type CashedFile struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	File      bool              `json:"file"`
	Public    bool              `json:"public"`
	Mime      string            `json:"mime,omitempty"`
	Grant     []string          `json:"grant"`
	Size      int               `json:"size,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt *time.Time        `json:"created"`
	Data      []byte            `json:"data,omitempty"`
	User      string            `json:"user"`
}