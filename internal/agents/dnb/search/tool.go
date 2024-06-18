package search

import (
	"context"

	"github.com/struki84/dnbclient"
)

type SearchTool struct {
	Client   *dnbclient.Client
	APIToken string
}

func NewSearch(apiToken string) *SearchTool {
	serchTool := &SearchTool{
		APIToken: apiToken,
	}

	c, err := dnbclient.NewClient(
		dnbclient.WithAPIKey(serchTool.APIToken),
	)

	if err != nil {
		return nil

	}

	serchTool.Client = c

	return serchTool
}

func (client *SearchTool) Name() string {
	return ""
}

func (client *SearchTool) Description() string {
	return ""
}

func (client *SearchTool) Call(ctx context.Context, input string) (string, error) {
	return "", nil
}
