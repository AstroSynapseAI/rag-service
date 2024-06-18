package api

import "github.com/tmc/langchaingo/llms"

type ClientOptions func(*APITool)

func WithActiveLLM(llm llms.Model) ClientOptions {
	return func(apiClient *APITool) {
		apiClient.ActiveLLM = llm
	}
}

func WithApiDocs(data string) ClientOptions {
	return func(apiClent *APITool) {
		apiClent.APIDocs = data
	}
}

func WithAPIToken(token string) ClientOptions {
	return func(apiClient *APITool) {
		apiClient.APIToken = token
	}
}
