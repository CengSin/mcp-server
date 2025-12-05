package tools

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
	"mcp/server/client"
	"mcp/server/dao"
)

func getUserBenefitRecordsTool() mcp.Tool {
	tool := mcp.NewTool("get_user_benefit_records",
		mcp.WithDescription("根据用户ID列表，查询他们是否领取了指定的栏目权限（如《脱水研报》、《早知道》）。脱水研报的id是581，早知道的id是679。"),
		mcp.WithArray("user_ids", mcp.Required(), mcp.WithNumberItems(mcp.Description("用户ID列表"))),
		mcp.WithArray("subject_ids", mcp.Required(), mcp.WithNumberItems(mcp.Description("栏目id列表"))),
	)
	return tool
}

type QueryUserBenefitRecords struct {
	UserIds    []int `json:"user_ids"`
	SubjectIds []int `json:"subject_ids"`
}

func getUserBenefitRecords(ctx context.Context, request mcp.CallToolRequest, params string) (*mcp.CallToolResult, error) {
	var args QueryUserBenefitRecords
	if err := json.Unmarshal([]byte(params), &args); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	var records []dao.ActivityFreeSubject
	if err := client.Mysql.Model(&dao.ActivityFreeSubject{}).Find(&records, "user_id in (?) and subject_id in (?)", args.UserIds, args.SubjectIds).Error; err != nil {
		return nil, err
	}

	return mcp.NewToolResultStructuredOnly(records), nil
}
