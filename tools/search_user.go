package tools

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"mcp/server/client"
	"mcp/server/dao"
	"time"
)

func getSearchUserTool() mcp.Tool {
	tool := mcp.NewTool("search_users",
		mcp.WithDescription("根据自然语言查询用户数据库,支持根据时间范围查询。"),
		mcp.WithInputSchema[SearchUserReq](),
		mcp.WithOutputSchema[[]*User](),
	)
	return tool
}

type SearchUserReq struct {
	StartTime *time.Time `json:"start_time" jsonschema_description:"查询开始时间, RFC3339 timestamp, e.g. 2024-12-31T23:59:59+08:00"`
	EndTime   *time.Time `json:"end_time" jsonschema_description:"查询结束时间, RFC3339 timestamp, e.g. 2024-12-31T23:59:59+08:00"`
	OrderBy   string     `json:"order_by" jsonschema_description:"排序字段，目前支持根据创建时间(created_at)，最新活跃时间(last_active_at)排序"`
	Sort      string     `json:"sort" jsonschema_description:"排序规则，desc表示降序，asc表示升序"`
	Limit     int        `json:"limit" jsonschema_description:"查询数量"`
}

type User struct {
	Id           int64
	Username     string
	Email        string
	Mobile       string
	Nickname     string
	RealName     string
	CreatedAt    time.Time
	LastActiveAt time.Time
	DeletedAt    *time.Time
}

func (u *User) Eich(u1 *dao.UserModel) {
	u.Id = u1.Id
	u.Username = u1.Username.String
	u.Email = u1.Email.String
	u.Mobile = u1.Mobile.String
	u.Nickname = u1.Nickname.String
	u.RealName = u1.RealName.String
	u.CreatedAt = u1.CreatedAt
	u.LastActiveAt = u1.LastActiveAt
	u.DeletedAt = u1.DeletedAt
}

func searchUser(ctx context.Context, request mcp.CallToolRequest, sq SearchUserReq) ([]*User, error) {
	var result []*dao.UserModel
	db := client.Mysql.Model(&dao.UserModel{})

	if sq.StartTime != nil {
		db = db.Where("created_at >= ?", sq.StartTime)
	}

	if sq.EndTime != nil {
		db = db.Where("created_at <= ?", sq.EndTime)
	}

	if sq.OrderBy != "" {
		db = db.Order(sq.OrderBy + " " + sq.Sort)
	}

	if sq.Limit <= 0 || sq.Limit > 200 {
		sq.Limit = 200
	}

	if err := db.Limit(sq.Limit).Scan(&result).Error; err != nil {
		return nil, err
	}

	var users []*User
	for _, u := range result {
		us := User{}
		us.Eich(u)
		users = append(users, &us)
	}

	return users, nil
}
