package callbacks

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/schema"
)

// DefaultKeywords is map of the agents final out prefix keywords.
//
//nolint:all
var DefaultKeywords = []string{"Final Answer:", "Final:", "AI:"}

// nolint:all
type StreamHandler struct {
	callbacks.SimpleHandler
	egress          chan []byte
	Keywords        []string
	LastTokens      string
	KeywordDetected bool
	PrintOutput     bool

	//tmp fix
	ChainsActive   []string
	ChainsFinished []string
}

var _ callbacks.Handler = &StreamHandler{}

func NewStreamHandler(keywords ...string) *StreamHandler {
	if len(keywords) > 0 {
		DefaultKeywords = keywords
	}

	return &StreamHandler{
		egress:         make(chan []byte),
		Keywords:       DefaultKeywords,
		ChainsActive:   make([]string, 0),
		ChainsFinished: make([]string, 0),
	}
}

func (handler *StreamHandler) GetEgress() chan []byte {
	return handler.egress
}

func (handler *StreamHandler) ReadFromEgress(ctx context.Context, callback func(ctx context.Context, chunk []byte)) {
	go func() {
		defer close(handler.egress)
		for data := range handler.egress {
			callback(ctx, data)
		}
	}()
}

func (handler *StreamHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	// ugly tmp fix
	chainID := "chain" + strconv.Itoa(len(handler.ChainsActive))
	handler.ChainsActive = append(handler.ChainsActive, chainID)

	if len(handler.ChainsActive) == 1 {
		jsonPayload := map[string]any{
			"step": "chain start",
		}
		jsonData, _ := json.Marshal(jsonPayload)
		handler.egress <- jsonData
	}
}

func (handler *StreamHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	// ugly tmp fix, need to do research into callbacks again

	finishedChainID := "chain" + strconv.Itoa(len(handler.ChainsFinished))
	handler.ChainsFinished = append(handler.ChainsFinished, finishedChainID)

	if len(handler.ChainsFinished) == len(handler.ChainsActive) {
		jsonPayload := map[string]any{
			"step": "chain end",
		}
		jsonData, _ := json.Marshal(jsonPayload)
		handler.egress <- jsonData
	}
}

func (handler *StreamHandler) HandleAgentFinish(_ context.Context, finish schema.AgentFinish) {
	jsonPayload := map[string]any{
		"step": "agent finish",
	}
	jsonData, _ := json.Marshal(jsonPayload)
	handler.egress <- jsonData
}

func (handler *StreamHandler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	jsonPayload := map[string]any{
		"step":  "agent action",
		"agent": action.Tool,
	}
	jsonData, _ := json.Marshal(jsonPayload)
	handler.egress <- jsonData
}

func (handler *StreamHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	chunkStr := string(chunk)
	handler.LastTokens += chunkStr

	// Buffer the last few chunks to match the longest keyword size
	longestSize := len(handler.Keywords[0])
	for _, k := range handler.Keywords {
		if len(k) > longestSize {
			longestSize = len(k)
		}
	}

	if len(handler.LastTokens) > longestSize {
		handler.LastTokens = handler.LastTokens[len(handler.LastTokens)-longestSize:]
	}

	// Check for keywords
	for _, k := range DefaultKeywords {
		if strings.Contains(handler.LastTokens, k) {
			handler.KeywordDetected = true
		}
	}

	// Check for colon and set print mode.
	if handler.KeywordDetected && chunkStr != ":" {
		handler.PrintOutput = true
	}

	// Print the final output after the detection of keyword.
	if handler.PrintOutput {
		jsonPayload := map[string]any{
			"step": "final output",
			"msg":  string(chunk),
		}
		jsonData, _ := json.Marshal(jsonPayload)
		handler.egress <- jsonData
	}
}
