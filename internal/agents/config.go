package agents

import (
	"fmt"
	"github.com/AstroSynapseAI/app-service/models"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type AgentConfig interface {
	GetAgentName() string
	GetAgentSlug() string
	GetAgentLLM() llms.Model
	GetAgentConfig() string
	GetAgentPrimer() string
	IsAgentPublic() bool
	IsAgentActive() bool
}

var _ AgentConfig = (*ActiveAgent)(nil)

// Active Agent Config
type ActiveAgent struct {
	Avatar      models.Avatar
	ActiveAgent models.ActiveAgent
}

func NewActiveAgent(avatar models.Avatar, activeAgent models.ActiveAgent) *ActiveAgent {
	return &ActiveAgent{
		Avatar:      avatar,
		ActiveAgent: activeAgent,
	}
}

func (cnf *ActiveAgent) GetAgentName() string {
	return cnf.ActiveAgent.Agent.Name
}

func (cnf *ActiveAgent) GetAgentSlug() string {
	return cnf.ActiveAgent.Agent.Slug
}

func (cnf *ActiveAgent) GetAgentLLM() llms.Model {
	agentLLM := cnf.ActiveAgent.LLM
	activeLLMs := cnf.Avatar.ActiveLLMs

	var activeLLM models.ActiveLLM

	for _, active := range activeLLMs {
		if active.LLM.ID == agentLLM.ID {
			activeLLM = active
		}
	}

	LLM, err := openai.New(
		openai.WithModel(activeLLM.LLM.Slug),
		openai.WithToken(activeLLM.Token),
	)

	if err != nil {
		fmt.Println("Error loading LLM:", err)
		return nil
	}
	return LLM
}

func (cnf *ActiveAgent) GetAgentConfig() string {
	return cnf.ActiveAgent.Config
}

func (cnf *ActiveAgent) GetAgentPrimer() string {
	return cnf.ActiveAgent.Primer
}

func (cnf *ActiveAgent) IsAgentPublic() bool {
	return cnf.ActiveAgent.IsPublic
}

func (cnf *ActiveAgent) IsAgentActive() bool {
	return cnf.ActiveAgent.IsActive
}
