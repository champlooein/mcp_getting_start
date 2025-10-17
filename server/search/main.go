package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	r := gin.Default()
	r.POST("/mcp", gin.WrapH(server.NewStreamableHTTPServer(NewMCPServer())))
	r.GET("/test", gin.WrapF(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello, world"))
	}))

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func NewMCPServer() *server.MCPServer {
	s := server.NewMCPServer("search_mcp", "0.0.2", server.WithToolCapabilities(true))
	s.AddTool(defaultSearchTool.getTavilySearchMCPTool())

	return s
}
