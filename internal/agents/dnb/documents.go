package dnb

import (
	"context"
)

type DocummentTool struct {
}

func NewDocummentTool() *DocummentTool {
	return &DocummentTool{}
}

func (DocummentTool) Name() string {
	return "DNB Document Tool"
}

func (DocummentTool) Description() string {
	return "DNB Document Tool"
}

func (DocummentTool) Call(ctx context.Context, input string) (string, error) {
	return "", nil
}
