package library

import (
	"context"
	"errors"
	"fmt"
	"time"

	asaiTools "github.com/AstroSynapseAI/rag-service/internal/tools"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

var _ tools.Tool = &LibraryAgent{}

type LibraryAgent struct {
	Memory   schema.Memory
	Primer   string
	LLM      llms.Model
	Executor *agents.Executor
}

func NewLibraryAgent(options ...LibraryAgentOptions) (*LibraryAgent, error) {
	libraryAgent := &LibraryAgent{
		Memory: memory.NewSimple(),
	}

	for _, option := range options {
		option(libraryAgent)
	}

	if libraryAgent.LLM == nil {
		return nil, errors.New("llm is required")
	}

	libraryTools := []tools.Tool{}

	promptTmplt := prompts.PromptTemplate{
		Template:       libraryAgent.Primer,
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		InputVariables: []string{"input", "agent_scratchpad", "today"},
		PartialVariables: map[string]interface{}{
			"today":             time.Now().Format("January 02, 2006"),
			"tool_names":        asaiTools.Names(libraryTools),
			"tool_descriptions": asaiTools.Descriptions(libraryTools),
			"history":           "",
		},
	}

	agent := agents.NewOneShotAgent(
		libraryAgent.LLM,
		libraryTools,
		agents.WithMemory(libraryAgent.Memory),
		agents.WithPrompt(promptTmplt),
		agents.WithMaxIterations(3),
	)

	libraryAgent.Executor = agents.NewExecutor(agent)

	return libraryAgent, nil
}

func (libraryAgent *LibraryAgent) Call(ctx context.Context, input string) (string, error) {
	fmt.Println("Library Agent Running...")

	return "Library Agent", nil
}

func (libraryAgent *LibraryAgent) Name() string {
	return "Library Agent"
}

func (libraryAgent *LibraryAgent) Description() string {
	return ""
}
