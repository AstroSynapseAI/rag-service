package library

import (
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
)

type LibraryAgentOptions func(agent *LibraryAgent)

func WithPrimer(primer string) LibraryAgentOptions {
	return func(agent *LibraryAgent) {
		agent.Primer = primer
	}
}

func WithLLM(llm llms.Model) LibraryAgentOptions {
	return func(agent *LibraryAgent) {
		agent.LLM = llm
	}
}

func WithExecutor(executor *agents.Executor) LibraryAgentOptions {
	return func(agent *LibraryAgent) {
		agent.Executor = executor
	}
}
