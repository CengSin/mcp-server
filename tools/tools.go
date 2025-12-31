package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	//s.AddTool(getContentMessagesTool(), mcp.NewTypedToolHandler(getContentMessages))
	//s.AddTool(getSearchArticleTool(), mcp.NewTypedToolHandler(searchArticle))
	s.AddTool(getSearchUserTool(), mcp.NewStructuredToolHandler(searchUser))
	s.AddTool(getUserBenefitRecordsTool(), mcp.NewTypedToolHandler(getUserBenefitRecords))
	s.AddTool(generateCsvTool(), mcp.NewTypedToolHandler(generateCsv))
}
