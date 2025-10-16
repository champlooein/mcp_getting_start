package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
)

func main() {
	var (
		s    = server.NewMCPServer("search_mcp", "0.0.1", server.WithToolCapabilities(false))
		tool = mcp.NewTool("search_tool", mcp.WithDescription("search internet content"), mcp.WithString("query", mcp.Required(), mcp.Description("To search for content")))
	)

	searchHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ret, err := Search(query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(ret), nil
	}

	s.AddTool(tool, searchHandler)
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func Search(query string) (string, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"query": query,
	})
	req, err := http.NewRequest(http.MethodPost, `https://api.tavily.com/search`, bytes.NewReader(requestBody))
	req.Header.Add("Authorization", "Bearer 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")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "launch http request err")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("http_status_code[%v] invalid", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "resp http body err")
	}

	type SearchResults struct {
		Results []struct {
			Content string `json:"content"`
		} `json:"results"`
	}
	searchResults := &SearchResults{}
	err = json.Unmarshal(respBody, searchResults)
	if err != nil {
		return "", errors.Wrap(err, "json.Unmarshal search result err")
	}

	ret := make([]string, 0)
	for _, result := range searchResults.Results {
		ret = append(ret, result.Content)
	}

	return strings.Join(ret, "\n\n"), nil
}
