package engine

import (
	"fmt"
	"strings"

	"github.com/AstroSynapseAI/app-service/models"
	"github.com/AstroSynapseAI/app-service/repositories"
	"github.com/AstroSynapseAI/rag-service/internal/agents"
	"github.com/AstroSynapseAI/rag-service/internal/agents/dnb"
	"github.com/AstroSynapseAI/rag-service/internal/agents/email"
	"github.com/AstroSynapseAI/rag-service/internal/agents/search"
	"github.com/GoLangWebSDK/crud/database"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

type Config struct {
	DB     *database.Database
	Avatar models.Avatar
}

type AvatarConfig interface {
	LoadConfig(userID uint)
	GetDB() *database.Database
	GetAvatarName() string
	GetAvatarLLM() llms.Model
	GetAvatarPrimer() string
	GetAvatarMemorySize() int
	AvatarIsPublic() bool
	GetAgents() []tools.Tool
	GetTools() []tools.Tool
}

var _ AvatarConfig = (*Config)(nil)

func NewConfig(db *database.Database) *Config {
	return &Config{
		DB: db,
	}
}

func (cnf *Config) LoadConfig(avatarID uint) {
	avatars := repositories.NewAvatarsRepository(cnf.DB)

	avatar, err := avatars.Fetch(avatarID)
	if err != nil {
		fmt.Println("Error loading avatar:", err)
		return
	}

	cnf.Avatar = avatar
}

func (cnf *Config) GetDB() *database.Database {
	return cnf.DB
}

func (cnf *Config) GetAvatarLLM() llms.Model {
	avatarLLM := cnf.Avatar.LLM
	activeLLMs := cnf.Avatar.ActiveLLMs
	if len(activeLLMs) == 0 {
		fmt.Println("Avatar has no active LLMs")
		return nil
	}

	// extract active llm where activeLLM.llmID == avatarLLM.ID
	var activeLLM models.ActiveLLM
	for _, active := range activeLLMs {
		if active.LLM.ID == avatarLLM.ID {
			activeLLM = active
		}
	}

	LLM, err := loadActiveLLM(activeLLM)

	if err != nil {
		fmt.Println("Error loading Avatar LLM:", err)
		return nil
	}
	return LLM
}

func (cnf *Config) GetAvatarName() string {
	return cnf.Avatar.Name
}

func (cnf *Config) GetAvatarPrimer() string {
	return cnf.Avatar.Primer
}

func (cnf *Config) GetAvatarMemorySize() int {
	return 4048
}

func (cnf *Config) AvatarIsPublic() bool {
	return cnf.Avatar.IsPublic
}

func (cnf *Config) GetAgents() []tools.Tool {
	activeAgents := cnf.Avatar.ActiveAgents
	loadedAgents := []tools.Tool{}

	for _, activeAgent := range activeAgents {
		agent := agents.NewActiveAgent(cnf.Avatar, activeAgent)

		if agent.GetAgentSlug() == "search-agent" && agent.IsAgentActive() {
			searchAgent, err := search.NewSearchAgent(
				search.WithPrimer(agent.GetAgentPrimer()),
				search.WithLLM(agent.GetAgentLLM()),
				search.WithConfig(agent.GetAgentConfig()),
			)

			if err != nil {
				fmt.Println("Error loading search agent:", err)
				return nil
			}

			loadedAgents = append(loadedAgents, searchAgent)
		}

		if agent.GetAgentSlug() == "email-agent" && agent.IsAgentActive() {
			emailAgent, err := email.NewEmailAgent(
				email.WithPrimer(agent.GetAgentPrimer()),
				email.WithLLM(agent.GetAgentLLM()),
				email.WithConfig(agent.GetAgentConfig()),
			)

			if err != nil {
				fmt.Println("Error loading email agent:", err)
				return nil
			}

			loadedAgents = append(loadedAgents, emailAgent)
		}

		if agent.GetAgentSlug() == "dnb-search-agent" && agent.IsAgentActive() {
			dnbAgent, err := dnb.NewDNBAgent(
				dnb.WithPrimer(agent.GetAgentPrimer()),
				dnb.WithLLM(agent.GetAgentLLM()),
				dnb.WithConfig(agent.GetAgentConfig()),
				dnb.WithActiveAgentID(agent.ActiveAgent.ID),
			)

			if err != nil {
				fmt.Println("Error loading dnb agent:", err)
				return nil
			}

			loadedAgents = append(loadedAgents, dnbAgent)
		}
	}
	return loadedAgents
}

func (cnf *Config) GetTools() []tools.Tool {
	// activeTools := cnf.Avatar.ActiveTools
	var loadedTools []tools.Tool

	return loadedTools
}

func loadActiveLLM(activeLLM models.ActiveLLM) (llms.Model, error) {
	var LLM llms.Model
	var err error

	llmProvider := strings.ToLower(activeLLM.LLM.Provider)

	if llmProvider == "mistral" {
		LLM, err = mistral.New(
			mistral.WithAPIKey(activeLLM.Token),
			mistral.WithModel(activeLLM.LLM.Slug),
		)
	}

	if llmProvider == "openai" {
		LLM, err = openai.New(
			openai.WithToken(activeLLM.Token),
			openai.WithModel(activeLLM.LLM.Slug),
		)
	}

	return LLM, err
}
