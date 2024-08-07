package tools

import (
	"github.com/AstroSynapseAI/asai-service/models"
)

type ToolConfig interface {
	GetName() string
	GetSlug() string
	GetToken() string
	GetConfig() string
	IsPublic() bool
	IsActive() bool
}

// Active Tool Config
type ActiveTool struct {
	Avatar     models.Avatar
	activeTool models.ActiveTool
}

var _ ToolConfig = (*ActiveTool)(nil)

func NewActiveTool(avatar models.Avatar, tool models.ActiveTool) *ActiveTool {
	return &ActiveTool{
		activeTool: tool,
		Avatar:     avatar,
	}
}

func (cnf *ActiveTool) GetName() string {
	return cnf.activeTool.Tool.Name
}

func (cnf *ActiveTool) GetSlug() string {
	return cnf.activeTool.Tool.Slug
}

func (cnf *ActiveTool) GetToken() string {
	return cnf.activeTool.Token
}

func (cnf *ActiveTool) GetConfig() string {
	return ""
}

func (cnf *ActiveTool) IsPublic() bool {
	return cnf.activeTool.IsPublic
}

func (cnf *ActiveTool) IsActive() bool {
	return cnf.activeTool.IsActive
}
