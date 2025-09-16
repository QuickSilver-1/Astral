package filescontroller

import "astral/internal/domain/file"

type docMeta struct {
	Name   string   `json:"name" binding:"required"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token" binding:"required"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

type uploadDataResponse struct {
	Json any    `json:"json,omitempty"`
	File string `json:"file,omitempty"`
}

type filesDataResponse struct {
	Docs []file.File `json:"docs"`
}

type deleteFileResponse struct {
	Response struct {
		ID bool `json:"file_id"`
	} `json:"response"`
}