package documents

import (
	"github.com/AstroSynapseAI/asai-service/models"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
)

type DocuemntsToolOption func(*DocumentsTool)

func WithRootPath(path string) DocuemntsToolOption {
	return func(tool *DocumentsTool) {
		tool.RootPath = path
	}
}

type SearchToolOption func(*SearchTool)

func WithDocuments(docs []models.Document) SearchToolOption {
	return func(tool *SearchTool) {
		tool.LoadedDocs = docs
	}
}

func WithEmbedder(embedder embeddings.Embedder) SearchToolOption {
	return func(tool *SearchTool) {
		tool.Embedder = embedder
	}
}

func WithModel(model llms.Model) SearchToolOption {
	return func(tool *SearchTool) {
		tool.Model = model
	}
}
