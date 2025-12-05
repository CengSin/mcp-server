package dao

import (
	"database/sql"
	"time"
)

type UserModel struct {
	Id            int64
	Username      sql.NullString `gorm:"size:63;unique_index"`
	Email         sql.NullString `gorm:"size:63;unique_index"`
	Mobile        sql.NullString `gorm:"size:15;unique_index"`
	PasswordSalt  []byte         `gorm:"size:64"`
	PasswordHash  []byte         `gorm:"size:64"`
	WeiboId       sql.NullString `gorm:"size:63;unique_index"`
	WeixinUnionId sql.NullString `gorm:"size:63;unique_index"`
	QqOpenId      sql.NullString `gorm:"size:63;unique_index"`
	OAuth0Id      sql.NullString `gorm:"size:63;unique_index"`
	OAuth1Id      sql.NullString `gorm:"size:63;unique_index"`
	// Only store appToken because webTokens are short lived
	AppToken      sql.NullString `gorm:"size:127"`
	Manufacturer  sql.NullString `gorm:"size:127"`
	PushProvider  sql.NullString `gorm:"size:16"`
	PushToken     sql.NullString `gorm:"size:255"`
	LastDeviceId  sql.NullString `gorm:"size:127"` // 用于追踪设备有没有注册过
	ClientVersion sql.NullString `gorm:"size:32"`  // 记录客户端版本号
	Nickname      sql.NullString `gorm:"size:63"`
	RealName      sql.NullString `gorm:"size:63"`
	Portrait      sql.NullString `gorm:"size:255"`
	BannedUntil   *time.Time
	CreatedAt     time.Time
	LastActiveAt  time.Time `gorm:"index"`
	DeletedAt     *time.Time
	PlatformName  string `gorm:"index"`
	XgbChannel    string `gorm:"index"`

	QqUnionId sql.NullString `gorm:"size:63;index"`

	AppleUserIdentifier sql.NullString `gorm:"size:100;unique_index"`
}

func (_ *UserModel) TableName() string {
	return "user_users"
}
