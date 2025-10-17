package search

import (
	"context"
	"sync"

	mcp_client "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

var (
	clientMap sync.Map
	mu        sync.Mutex

	queryKey = "query"
	topicKey = "topic"

	tavilySearchToolName = "tavily_search"
)

type Client interface {
	TavilySearch(ctx context.Context, query string, topic *string) (string, error)
}

type client struct {
	mcpClient *mcp_client.Client
}

func NewClient(ctx context.Context, baseURL string) (Client, error) {
	if c, ok := clientMap.Load(baseURL); ok {
		return c.(Client), nil
	}

	mu.Lock()
	defer mu.Unlock()
	if c, ok := clientMap.Load(baseURL); ok {
		return c.(Client), nil
	}

	httpTransport, err := transport.NewStreamableHTTP(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP transport")
	}

	mcpClient := mcp_client.NewClient(httpTransport)
	if err = mcpClient.Start(ctx); err != nil {
		defer httpTransport.Close()
		return nil, errors.Wrap(err, "failed to start MCP client")
	}
	if _, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		defer httpTransport.Close()
		return nil, errors.Wrap(err, "failed to initialize MCP client")
	}

	c := &client{mcpClient: mcpClient}
	clientMap.Store(baseURL, c)
	return c, nil
}

func (c *client) TavilySearch(ctx context.Context, query string, topic *string) (string, error) {
	args := map[string]string{queryKey: query}
	if topic != nil {
		args[topicKey] = *topic
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      tavilySearchToolName,
			Arguments: args,
		},
	}
	resp, err := c.mcpClient.CallTool(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "failed to call tavily search tool")
	}
	if resp.IsError {
		return "", errors.New(resp.Content[0].(mcp.TextContent).Text)
	}

	return resp.Content[0].(mcp.TextContent).Text, nil
}
