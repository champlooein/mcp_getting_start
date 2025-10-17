package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type searchTool struct{}

var (
	defaultSearchTool = &searchTool{}
)

const (
	tavilySearchResultTpl = "Summary of search results:\n{{.answer}}\n\nSearch result details:{{range .content}}\n- {{.}}{{end}}\n"
	defaultSearchTopic    = "general"

	tavilySearchToolName = "tavily_search"

	queryKey = "query"
	topicKey = "topic"
)

func (t searchTool) getTavilySearchMCPTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool(
		tavilySearchToolName,
		mcp.WithDescription("A search engine API optimized for Large Language Model (LLM) and Retrieval-Augmented Generation (RAG) applications, designed to efficiently and quickly provide real-time, accurate, and relevant web search results to enhance the information acquisition and processing capabilities of AI agents."),
		mcp.WithString("query", mcp.Required(), mcp.Description("To search for content")),
		mcp.WithString("topic", mcp.Description("The category of the search. Available options: general, news, finance. news is useful for retrieving real-time updates, particularly about politics, sports, and major current events covered by mainstream media sources. general is for broader, more general-purpose searches that may include a wide range of sources.")),
	)

	tavilySearchHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString(queryKey)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ret, err := t.tavilySearch(query, request.GetString(topicKey, defaultSearchTopic))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(ret), nil
	}

	return tool, tavilySearchHandler
}

func (t searchTool) tavilySearch(query, topic string) (string, error) {
	var (
		tavilyApiKey   = os.Getenv("TAVILY_API_KEY")
		requestBody, _ = json.Marshal(map[string]interface{}{
			"query":          query,
			"topic":          topic,
			"max_results":    10,
			"include_answer": "advanced",
		})
	)
	if tavilyApiKey == "" {
		return "", errors.New("tavily api key is empty")
	}

	req, err := http.NewRequest(http.MethodPost, `https://api.tavily.com/search`, bytes.NewReader(requestBody))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tavilyApiKey))
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

	type Result struct {
		Content string `json:"content"`
	}
	type SearchResults struct {
		Answer  string   `json:"answer"`
		Results []Result `json:"results"`
	}
	searchResults := &SearchResults{}
	err = json.Unmarshal(respBody, searchResults)
	if err != nil {
		return "", errors.Wrap(err, "json.Unmarshal search result err")
	}

	return t.formatTavilySearchResult(searchResults.Answer, lo.Map(searchResults.Results, func(result Result, index int) string { return result.Content }))
}

func (t searchTool) formatTavilySearchResult(ans string, content []string) (string, error) {
	parsedTmpl, err := template.New("template").Option("missingkey=error").Parse(tavilySearchResultTpl)
	if err != nil {
		return "", err
	}
	sb := new(strings.Builder)
	err = parsedTmpl.Execute(sb, map[string]any{"answer": ans, "content": content})
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
