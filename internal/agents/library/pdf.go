package library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/AstroSynapseAI/asai-service/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/textsplitter"
)

type PDFAgent struct {
	File       *bytes.Reader
	Splitter   textsplitter.RecursiveCharacter
	Embedder   embeddings.Embedder
	Model      llms.Model
	avatarDocs []models.Document
}

func NewPDFAgent(options ...PDFAgentOptions) (*PDFAgent, error) {
	pdfAgent := &PDFAgent{}

	pdfAgent.Splitter = textsplitter.NewRecursiveCharacter()
	pdfAgent.Splitter.ChunkSize = 500
	pdfAgent.Splitter.ChunkOverlap = 50

	return pdfAgent, nil
}

func (agent *PDFAgent) Name() string {
	return "PDF reader tool."
}

func (agent *PDFAgent) Description() string {
	str := "Enables your avatar to read PDF files. \n\n Avaliable files: \n"

	for _, doc := range agent.avatarDocs {
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

	err = agent.loadFile(toolInput.File)
	if err != nil {
		return "", err
	}

	PDFLoader := documentloaders.NewPDF(agent.File, agent.File.Size())

	docs, err := PDFLoader.LoadAndSplit(ctx, agent.Splitter)
	if err != nil {
		return "", err
	}

	// store, err := pinecone.New(
	// 	pinecone.WithNameSpace("asai"),
	// 	pinecone.WithEmbedder(agent.Embedder),
	// )
	//
	// if err != nil {
	// 	return "", err
	// }
	//
	// _, err = store.AddDocuments(ctx, docs)
	// if err != nil {
	// 	return "", err
	// }
	//
	// docs, err = store.SimilaritySearch(
	// 	ctx,
	// 	input,
	// 	1,
	// 	vectorstores.WithScoreThreshold(0.5),
	// )
	//
	// if err != nil {
	// 	return "", err
	// }

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

func (agent *PDFAgent) loadFile(file string) error {
	var loadedDoc models.Document

	for _, doc := range agent.avatarDocs {
		if doc.Name == file {
			loadedDoc = doc
			break
		}
	}

	if loadedDoc.Name == "" {
		return fmt.Errorf("file not found: %s", file)
	}

	sess, err := session.NewSession(agent.createAWSSession())
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
	agent.File = bytes.NewReader(content)

	return nil
}

func (agent *PDFAgent) createAWSSession() *aws.Config {

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
