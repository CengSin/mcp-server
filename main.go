package main

import (
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"log"
	"mcp/server/client"
	"mcp/server/config"
	"mcp/server/tools"
	"os"
)

func main() {
	os.Setenv("HTTP_PROXY", "127.0.0.1:7890")
	os.Setenv("HTTPS_PROXY", "127.0.0.1:7890")

	config.Parse()

	cfg := config.Cfg
	client.InitMysql(cfg.Mysql)
	client.InitQdrant(cfg.Qdrant)
	client.InitLLMs(cfg.OpenAI)
	client.InitMinIO(cfg.MinIO)
	defer client.Close()

	mcpServer := server.NewMCPServer("rag_finance_news_tools", "1.0.0")

	tools.RegisterTools(mcpServer)

	httpServer := server.NewStreamableHTTPServer(mcpServer)

	log.Println("Starting StreamableHTTP server on :8085")
	if err := httpServer.Start(":8085"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
