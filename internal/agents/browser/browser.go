package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/AstroSynapseAI/rag-service/internal/tools/scraper"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/tools"
)

var _ tools.Tool = &BrowserAgent{}

const (
	ErrParsingInput       = "Browser Agent failed to parse input"
	ErrScraping           = "Browser Agent failed to scrape web"
	ErrLoadingDocuments   = "Browser Agent failed to load web into documents"
	ErrSummarising        = "Browser Agent failed to summarise web"
	SummarisationTemplate = `
		Please write a detailed report of the following website and its pages that will not exceed 4048 tokens:

		"{{.context}}"

		If query is provided, focus on the content related to the query. 

		Query: {{.query}}

		Structure the report in the following format:

		WEBSITE SUMMARY:
		[Place the summary of the entire website here]

		PAGE SUMMARIES:
		- [Page 1 Title]: [Summary of Page 1]
		- [Page N Title]: [Summary of Page N]
		...(Create a summary for every sub-page on the website)

		LINK INDEX:
		- Link 1: [Description of Link 1]
		- Link N: [Description of Link N]
		...(Depending on relevance, you can add none or N number of links)

		FINAL THOUGHTS:
		[Place any final thoughts or a concluding summary here]`
)

type BrowserAgent struct {
	llm     *openai.LLM
	Scraper *scraper.Scraper
	Memory  schema.Memory
}

func New(options ...BrowserAgentOptions) (*BrowserAgent, error) {
	var err error
	browserAgent := &BrowserAgent{
		Memory: memory.NewSimple(),
	}

	for _, option := range options {
		option(browserAgent)
	}

	browserAgent.llm, err = openai.New(openai.WithModel("gpt-4"))
	if err != nil {
		return nil, err
	}

	browserAgent.Scraper, err = scraper.New()
	if err != nil {
		return nil, err
	}

	return browserAgent, nil
}

func (agent *BrowserAgent) Name() string {
	return "Web Browser Agent"
}

func (agent *BrowserAgent) Description() string {
	return `
		Web Browser Agent is an agent specialized in scraping and reading the web pages.
		It will read and summarize the content of the web page and return it as output.
		
		Input should be a json string containing the url of the web page a query string.  
		The query string, if set, will tell the agent what to search for within the web page, 
		leave blank for a general summary.

		Input JSON Format:
		{
			"url": "[url of the web page]", 
			"query": "[query string]"
		}
	`
}

func (agent *BrowserAgent) Call(ctx context.Context, input string) (string, error) {
	jsonString, err := extractJSON(input)
	if err != nil {
		return ErrParsingInput, nil
	}

	var jsonInput struct {
		URL   string `json:"url"`
		Query string `json:"query"`
	}

	err = json.Unmarshal([]byte(jsonString), &jsonInput)
	if err != nil {
		return ErrParsingInput, nil
	}

	webDocuments, err := agent.loadWebContent(ctx, jsonInput.URL)
	if err != nil {
		return ErrLoadingDocuments, nil
	}

	llmChain := chains.NewLLMChain(agent.llm, prompts.NewPromptTemplate(
		SummarisationTemplate, []string{"context", "query"},
	))

	summaryChain := chains.NewStuffDocuments(llmChain)
	summary, err := chains.Call(
		ctx,
		summaryChain,
		map[string]any{
			"input_documents": webDocuments,
			"query":           jsonInput.Query,
		},
	)

	if err != nil {
		return ErrSummarising, nil
	}

	response := summary["text"].(string)
	return response, nil
}

func (agent *BrowserAgent) loadWebContent(ctx context.Context, url string) ([]schema.Document, error) {
	webContent, err := agent.Scraper.Call(ctx, url)
	if err != nil {
		return []schema.Document{}, err
	}

	webContentReader := strings.NewReader(webContent)

	loader := documentloaders.NewText(webContentReader)
	if err != nil {
		return []schema.Document{}, err
	}

	spliter := textsplitter.NewTokenSplitter()
	spliter.ChunkSize = 7500
	spliter.ChunkOverlap = 1024
	spliter.ModelName = "gpt-4"

	webDocuments, err := loader.LoadAndSplit(ctx, spliter)
	if err != nil {
		return []schema.Document{}, err
	}

	return webDocuments, nil
}

func extractJSON(input string) (string, error) {
	re := regexp.MustCompile(`(?s)\{.*\}`)
	jsonString := re.FindString(input)
	if jsonString == "" {
		return "", fmt.Errorf("invalid JSON object found in Browser Agent input")
	}

	return jsonString, nil
}
