package tools

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"mcp/server/client"
	"mcp/server/dao"
	"mcp/server/util"
	"strings"
	"time"
)

func getContentMessagesTool() mcp.Tool {
	tool := mcp.NewTool("search_content_messages",
		mcp.WithDescription(`
MySQL 精确查询工具：
- 适合：按时间排序、获取最新文章、按字段过滤（如 source_id、日期、分类）
- 不适用：意图模糊、纯自然语言语义理解类问题（如 “有哪些讲AI趋势的文章？”）
如果用户请求涉及 “最新文章”、“按时间排序”、“topN 列表”、“字段条件”，必须优先使用此工具。
`),
		mcp.WithString("keyword", mcp.Description("结构化关键词（文章标题/内容中的字段），非语义问题")),
		mcp.WithString("start_time", mcp.Description("开始时间，格式为2006-01-02 15:04:05，最早可到2024-01-01 00:00:00")),
		mcp.WithString("end_time", mcp.Description("结束时间，格式为2006-01-02 15:04:05，最晚可到当前时间")),
		mcp.WithString("order_by", mcp.Description("排序字段，如 created_at")),
		mcp.WithString("order_direction", mcp.Description("排序方向，asc 或 desc")),
		mcp.WithNumber("limit", mcp.Description("返回结果数量，默认为 5，最大不超过100")))
	return tool
}

type getContentMessagesReq struct {
	Keyword        string `json:"keyword"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	Limit          int    `json:"limit"`
	OrderBy        string `json:"order_by"`
	OrderDirection string `json:"order_direction"`
}

func getContentMessages(ctx context.Context, request mcp.CallToolRequest, searchReq getContentMessagesReq) (*mcp.CallToolResult, error) {
	tx := client.Mysql.Model(&dao.ContentMessage{})

	if searchReq.StartTime != "" && searchReq.EndTime != "" {
		startTime, _ := time.ParseInLocation(time.DateTime, searchReq.StartTime, util.Loc)
		endTime, _ := time.ParseInLocation(time.DateTime, searchReq.EndTime, util.Loc)
		tx = tx.Where("created_at between ? and ?", startTime.UTC(), endTime.UTC())
	}

	if searchReq.Keyword != "" {
		tx = tx.Where("content LIKE ?", "%"+strings.TrimSpace(searchReq.Keyword)+"%")
	}

	if searchReq.OrderBy != "" {
		tx = tx.Order(searchReq.OrderBy + " " + searchReq.OrderDirection)
	}

	var result []dao.ContentMessage
	if err := tx.Limit(searchReq.Limit).Find(&result).Error; err != nil {
		return nil, err
	}

	return mcp.NewToolResultStructuredOnly(result), nil
}
