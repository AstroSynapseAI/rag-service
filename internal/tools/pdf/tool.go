package pdf

import (
	"context"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"

	"github.com/google/uuid"
)

var _ tools.Tool = PDFTool{}

type PDFTool struct {
	File     *os.File
	Splitter textsplitter.RecursiveCharacter
}

func NewTool() (PDFTool, error) {
	pdfTool := PDFTool{}

	pdfTool.Splitter = textsplitter.NewRecursiveCharacter()
	pdfTool.Splitter.ChunkSize = 500
	pdfTool.Splitter.ChunkOverlap = 50

	return pdfTool, nil
}

func (tool PDFTool) Name() string {
	return "PDF reader tool."
}

func (tool PDFTool) Description() string {
	return "Enables your avatar to read PDF files."
}

func (tool PDFTool) Call(ctx context.Context, input string) (string, error) {
	var err error

	path := "test.pdf"

	tool.File, err = os.Open(path)
	if err != nil {
		return "", err
	}

	defer tool.File.Close()

	fileInfo, err := tool.File.Stat()
	if err != nil {
		return "", err
	}

	PDFLoader := documentloaders.NewPDF(tool.File, fileInfo.Size())

	docs, err := PDFLoader.LoadAndSplit(ctx, tool.Splitter)
	if err != nil {
		return "", err
	}

	llm, err := openai.New(openai.WithToken("token"))
	if err != nil {
		return "", err
	}

	embedder, err := embeddings.NewEmbedder(llm)

	store, err := pinecone.New(
		ctx,
		pinecone.WithNameSpace(uuid.New().String()),
		pinecone.WithProjectName("fd4e2b9"),
		pinecone.WithAPIKey("65ae7457-8f3d-4b23-a54e-d19b827ab218"),
		pinecone.WithEnvironment("us-west1-gcp-free"),
		pinecone.WithIndexName("reading-test"),
		pinecone.WithEmbedder(embedder),
	)

	if err != nil {
		return "", err
	}

	_, err = store.AddDocuments(ctx, docs)
	if err != nil {
		return "", err
	}

	docs, err = store.SimilaritySearch(
		ctx,
		input,
		1,
		vectorstores.WithScoreThreshold(0.5),
	)

	if err != nil {
		return "", err
	}

	QAChain := chains.LoadStuffQA(llm)

	answer, err := chains.Call(ctx, QAChain, map[string]any{
		"input_documents": docs,
		"question":        input,
	})

	if err != nil {
		return "", err
	}

	response := answer["text"].(string)

	return response, nil
}

func (tool PDFTool) loadFile(ctx context.Context, path string) error {

	return nil
}
