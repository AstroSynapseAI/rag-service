package library

import (
	"github.com/AstroSynapseAI/asai-service/models"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
)

type LibraryAgentOptions func(agent *LibraryAgent)
type PDFAgentOptions func(agent *PDFAgent)

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

func WithEmbedder(embedder embeddings.Embedder) PDFAgentOptions {
	return func(agent *PDFAgent) {
		agent.Embedder = embedder
	}
}

func WithModel(model llms.Model) PDFAgentOptions {
	return func(agent *PDFAgent) {
		agent.Model = model
	}
}

func WithDocuments(docs []models.Document) PDFAgentOptions {
	return func(agent *PDFAgent) {
		agent.avatarDocs = docs
	}
}
