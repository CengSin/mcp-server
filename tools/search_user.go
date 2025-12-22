package tools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/gorm"
	"mcp/server/client"
	"mcp/server/dao"
	"mcp/server/util"
)

func getSearchUserTool() mcp.Tool {
	tool := mcp.NewTool("search_users",
		mcp.WithDescription("根据自然语言查询用户数据库,支持根据时间范围查询。"),
		mcp.WithString("start_time", mcp.Description("开始时间，格式为2006-01-02 15:04:05")),
		mcp.WithString("end_time", mcp.Description("结束时间，格式为2006-01-02 15:04:05")),
		mcp.WithNumber("limit", mcp.Description("查询数量，默认为5")),
	)
	return tool
}

type SearchUserReq struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Limit     int    `json:"limit"`
}

func searchUser(ctx context.Context, request mcp.CallToolRequest, sq SearchUserReq) (*mcp.CallToolResult, error) {
	startTime, err := util.ParseTime(sq.StartTime)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("startTime Parse err, %s", err.Error())), nil
	}

	endTime, err := util.ParseTime(sq.EndTime)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("endTime Parse err, %s", err.Error())), nil
	}

	if sq.Limit == 0 {
		sq.Limit = 5
	}

	var result []*dao.UserModel
	if err = client.Mysql.Model(&dao.UserModel{}).Limit(sq.Limit).Find(&result, "created_at between ? and ?", startTime, endTime).Error; err != nil && gorm.ErrRecordNotFound != err {
		return mcp.NewToolResultError(fmt.Sprintf("query user data err, %s", err.Error())), nil
	}

	return mcp.NewToolResultStructuredOnly(result), nil
}
