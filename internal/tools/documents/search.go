package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/AstroSynapseAI/mar-mar-service/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"

	"github.com/google/uuid"
)

var _ tools.Tool = &SearchTool{}

type SearchTool struct {
	File       *bytes.Reader
	Splitter   textsplitter.RecursiveCharacter
	LoadedDocs []models.Document
	Embedder   embeddings.Embedder
	Model      llms.Model
}

func NewTool(options ...SearchToolOption) (*SearchTool, error) {
	searchTool := &SearchTool{}

	searchTool.Splitter = textsplitter.NewRecursiveCharacter()
	searchTool.Splitter.ChunkSize = 500
	searchTool.Splitter.ChunkOverlap = 50

	return searchTool, nil
}

func (tool SearchTool) Name() string {
	return "PDF reader tool."
}

func (tool SearchTool) Description() string {
	str := "Enables your avatar to read and search through PDF files. \n\n Avaliable files: \n"

	for _, doc := range tool.LoadedDocs {
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

func (tool SearchTool) Call(ctx context.Context, input string) (string, error) {
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
		return fmt.Sprintf("%v: %s", ErrInvalidInput, err), nil
	}

	err = tool.loadFile(toolInput.File)
	if err != nil {
		return "", err
	}

	PDFLoader := documentloaders.NewPDF(tool.File, tool.File.Size())

	docs, err := PDFLoader.LoadAndSplit(ctx, tool.Splitter)
	if err != nil {
		return "", err
	}

	store, err := pinecone.New(
		pinecone.WithNameSpace(uuid.New().String()),
		pinecone.WithEmbedder(tool.Embedder),
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

	QAChain := chains.LoadStuffQA(tool.Model)

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

func (tool SearchTool) loadFile(file string) error {
	var loadedDoc models.Document

	for _, doc := range tool.LoadedDocs {
		if doc.Name == file {
			loadedDoc = doc
			break
		}
	}

	if loadedDoc.Name == "" {
		return fmt.Errorf("file not found: %s", file)
	}

	sess, err := session.NewSession(tool.createAWSSession())
	if err != nil {
		return fmt.Errorf("error creating AWS session: %w", err)
	}

	svc := s3.New(sess)

	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("mar-mar"),
		Key:    aws.String(loadedDoc.Name),
	})
	if err != nil {
		return fmt.Errorf("error getting object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the entire content into a byte slice
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return fmt.Errorf("error reading S3 object: %w", err)
	}

	// Create a *bytes.Reader from the content
	tool.File = bytes.NewReader(content)

	return nil
}

func (ctrl SearchTool) createAWSSession() *aws.Config {

	if os.Getenv("ENVIRONMENT") == "LOCAL DEV" {
		return &aws.Config{
			Region: aws.String("eu-central-1"),
			Credentials: credentials.NewStaticCredentials(
				os.Getenv("ACCESS_KEY"),
				os.Getenv("SECRET_KEY"),
				"",
			),
		}
	}

	return &aws.Config{
		Region: aws.String("eu-central-1"),
	}
}
