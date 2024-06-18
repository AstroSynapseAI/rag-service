package browser

import (
	"github.com/tmc/langchaingo/schema"
)

type BrowserAgentOptions func(agent *BrowserAgent)

func WithMemory(memory schema.Memory) BrowserAgentOptions {
	return func(agent *BrowserAgent) {
		agent.Memory = memory
	}
}