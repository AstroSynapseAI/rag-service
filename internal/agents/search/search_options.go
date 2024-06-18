package search

import (
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type config struct {
	DDGIsActive    bool   `json:"ddg_is_active,omitempty"`
	GoogleAPIToken string `json:"google_api_token,omitempty"`
	GoogleIsActive bool   `json:"google_is_active,omitempty"`
	ExaAPIToken    string `json:"exa_api_token,omitempty"`
	ExaIsActive    bool   `json:"exa_is_active,omitempty"`
}

type SearchAgentOptions func(agent *SearchAgent)

func WithConfig(data string) SearchAgentOptions {
	return func(agent *SearchAgent) {
		var configData config

		err := json.Unmarshal([]byte(data), &configData)
		if err != nil {
			fmt.Println("Error decoding search onfig data:", err)
			return
		}
		agent.Config = configData
	}
}

func WithMemory(memory schema.Memory) SearchAgentOptions {
	return func(agent *SearchAgent) {
		agent.Memory = memory
	}
}

func WithPrimer(primer string) SearchAgentOptions {
	return func(agent *SearchAgent) {
		agent.Primer = primer
	}
}

func WithLLM(llm llms.Model) SearchAgentOptions {
	return func(agent *SearchAgent) {
		agent.LLM = llm
	}
}
