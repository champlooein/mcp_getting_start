package search_test

import (
	"context"
	"testing"

	"github.com/champlooein/mcp_getting_started/client/search"
)

func Test_TavilySearch(t *testing.T) {
	c, err := search.NewClient(context.Background(), "http://localhost:8080/mcp/")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ret, err := c.TavilySearch(context.Background(), "今天的日期是多少？", nil)
	if err != nil {
		t.Fatalf("TavilySearch failed: %v", err)
	}

	t.Logf("TavilySearch result: %s", ret)
}
