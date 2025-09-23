package types

import (
	"github.com/go-playground/validator/v10"
)

type ListDetails struct {
	ID        string `json:"id" validate:"required"`
	SourceURL string `json:"sourceUrl" validate:"required,url"`
	Schema    string `json:"schema"` // can be a URL or provided schema
}

type ListOfTokenLists struct {
	Timestamp  string        `json:"timestamp"`
	Version    Version       `json:"version"`
	TokenLists []ListDetails `json:"tokenLists"`
}

func (fd *ListDetails) Validate() error {
	validate := validator.New()
	return validate.Struct(fd)
}
