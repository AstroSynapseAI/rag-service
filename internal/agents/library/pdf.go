package library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/AstroSynapseAI/asai-service/models"
	"github.com/AstroSynapseAI/rag-service/utils/storage"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

type PDFAgent struct {
	Splitter textsplitter.RecursiveCharacter
	Embedder embeddings.Embedder
	Model    llms.Model
	Docs     []models.Document
	Storage  *storage.AWS
}

func NewPDFAgent(options ...PDFAgentOptions) (*PDFAgent, error) {
	pdfAgent := &PDFAgent{}

	pdfAgent.Splitter = textsplitter.NewRecursiveCharacter()
	pdfAgent.Splitter.ChunkSize = 500
	pdfAgent.Splitter.ChunkOverlap = 50

	pdfAgent.Storage = storage.NewAWSStorage()

	return pdfAgent, nil
}

func (agent *PDFAgent) Name() string {
	return "PDF reader tool."
}

func (agent *PDFAgent) Description() string {
	str := "Enables your avatar to read PDF files. \n\n Avaliable files: \n"

	for _, doc := range agent.Docs {
		str += str + "- " + doc.Name + "\n"
	}

	str += `The tool exepects input in JSON format with search query and filename.\n
	{
		"query": "Where did Simun work in 2015?",
		"file": "SimunStukanCV.pdf"
	}\n
	`

	return str
}

func (agent *PDFAgent) Call(ctx context.Context, input string) (string, error) {
	fmt.Println("PDF reader tool called")

	var toolInput struct {
		Query string `json:"query,omitempty"`
		File  string `json:"file,omitempty"`
	}

	re := regexp.MustCompile(`(?s)\{.*\}`)
	jsonString := re.FindString(input)

	err := json.Unmarshal([]byte(jsonString), &toolInput)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("%v: %s", "invalid input", err), nil
	}

	fileByte, err := agent.Storage.GetFile(toolInput.File)
	if err != nil {
		return "", err
	}

	file := bytes.NewReader(fileByte)
	PDFLoader := documentloaders.NewPDF(file, file.Size())
	docs, err := PDFLoader.LoadAndSplit(ctx, agent.Splitter)
	if err != nil {
		return "", err
	}

	QAChain := chains.LoadStuffQA(agent.Model)
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

func (agent *PDFAgent) search(ctx context.Context, docs []schema.Document, input string) ([]schema.Document, error) {

	store, err := pinecone.New(
		pinecone.WithNameSpace("asai"),
		pinecone.WithEmbedder(agent.Embedder),
	)

	if err != nil {
		return nil, err
	}

	_, err = store.AddDocuments(ctx, docs)
	if err != nil {
		return nil, err
	}

	docs, err = store.SimilaritySearch(
		ctx,
		input,
		1,
		vectorstores.WithScoreThreshold(0.5),
	)

	if err != nil {
		return nil, err
	}

	return docs, nil
}
