package dnb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AstroSynapseAI/asai-service/models"
	"github.com/AstroSynapseAI/asai-service/repositories"
	"github.com/AstroSynapseAI/rag-service/internal/agents/dnb/api"
	"github.com/AstroSynapseAI/rag-service/internal/agents/dnb/search"
	util "github.com/AstroSynapseAI/rag-service/internal/tools"
	"github.com/GoLangWebSDK/crud/database"
	"github.com/struki84/dnbclient"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

type DNBAgent struct {
	DB            *database.Database
	ActiveAgentID uint
	Primer        string
	LLM           llms.Model
	Executor      *agents.Executor
	Config        config
	Tools         []tools.Tool
}

var _ tools.Tool = &DNBAgent{}

func NewDNBAgent(options ...DNBAgentOptions) (*DNBAgent, error) {
	dnbAgent := &DNBAgent{}

	for _, option := range options {
		option(dnbAgent)
	}

	err := dnbAgent.validateAPIToken()
	if err != nil {
		return dnbAgent, err
	}

	apiTool := api.NewTool(
		api.WithActiveLLM(dnbAgent.LLM),
		api.WithApiDocs(dnbAgent.loadAPIDocs("")),
		api.WithAPIToken(dnbAgent.Config.DNBAPIToken),
	)

	searchTool := search.NewSearch(dnbAgent.Config.DNBAPIToken)

	docTool := NewDocummentTool()

	dnbAgent.Tools = []tools.Tool{
		docTool,
		apiTool,
		searchTool,
	}

	agent := agents.NewOneShotAgent(
		dnbAgent.LLM,
		dnbAgent.Tools,
		agents.WithPrompt(dnbAgent.loadTemplate()),
		agents.WithMaxIterations(5),
	)

	dnbAgent.Executor = agents.NewExecutor(agent)

	return dnbAgent, nil
}

func (dnbAgent DNBAgent) Name() string {
	return "DNB"
}

func (dnbAgent DNBAgent) Description() string {
	return "DNB agent"
}

func (dnbAgent DNBAgent) Call(ctx context.Context, input string) (string, error) {
	fmt.Println("DNB Agent Running...")

	response, err := chains.Run(ctx, dnbAgent.Executor, input)
	if err != nil {
		return "DNB Agent encountered an error: " + err.Error(), nil
	}

	return response, nil
}

func (dnbAgent *DNBAgent) loadTemplate() prompts.PromptTemplate {
	// load template
	return prompts.PromptTemplate{
		Template:       dnbAgent.Primer,
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		InputVariables: []string{"input", "agent_scratchpad", "today"},
		PartialVariables: map[string]interface{}{
			"today":             "2023-01-01",
			"tool_names":        util.Names(dnbAgent.Tools),
			"tool_descriptions": util.Descriptions(dnbAgent.Tools),
			"history":           "",
		},
	}
}

func (dnbAgent *DNBAgent) loadAPIDocs(file string) string {
	path := ""
	docs, err := os.ReadFile(path)
	if err != nil {
		log.Println("Error reading api docs: ", err)
	}

	return string(docs)
}

func (dnbAgent *DNBAgent) validateAPIToken() error {
	if dnbAgent.Config.DNBAPIKey == "" || dnbAgent.Config.DNBAPISecret == "" {
		return fmt.Errorf("DNB API key and secret are required")
	}

	dnbClient, err := dnbclient.NewClient(
		dnbclient.WithTokens(
			dnbAgent.Config.DNBAPIKey,
			dnbAgent.Config.DNBAPISecret,
		),
	)

	if err != nil {
		return err
	}

	if dnbAgent.Config.DNBAPIToken == "" || dnbAgent.expiredToken() {
		repo := repositories.NewAgentsRepository(dnbAgent.DB)

		apiToken, err := dnbClient.GetToken(context.Background())
		if err != nil {
			return err
		}

		dnbAgent.Config.DNBAPIToken = apiToken
		dnbAgent.Config.TokenAge = time.Now().Unix()

		jsonConfig, err := json.Marshal(dnbAgent.Config)
		if err != nil {
			return err
		}

		var agentData models.ActiveAgent
		agentData.ID = dnbAgent.ActiveAgentID
		agentData.Config = string(jsonConfig)

		_, err = repo.Active.Update(dnbAgent.ActiveAgentID, agentData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dnbAgent *DNBAgent) expiredToken() bool {
	return dnbAgent.Config.TokenAge < time.Now().Add(-24*time.Hour).Unix()
}
