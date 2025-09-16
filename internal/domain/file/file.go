package file

import (
	"io"
	"time"
)

type File struct {
	ID        string   		    `json:"id"`
	Name      string   		    `json:"name"`
	File      bool     		    `json:"file"`
	Public    bool     		    `json:"public"`
	Mime      string   		    `json:"mime,omitempty"`
	Grant     []string 		    `json:"grant"`
	Size      int			    `json:"size,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt *time.Time        `json:"created"`
	Reader 	  io.Reader			`json:"-"`
	User      string			`json:"-"`
}