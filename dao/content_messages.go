package dao

import (
	"database/sql"
	"time"
)

type ContentMessage struct {
	Id              int64
	AuthorId        int64
	EditorId        int64
	Style           string `gorm:"index"`
	IsPremium       bool   `sql:"DEFAULT:0" gorm:"index"`
	IsTrial         bool   `sql:"DEFAULT:0" gorm:"index"`
	SubscribeType   string `sql:"DEFAULT:0" gorm:"index"`
	IsPromotion     bool   `sql:"DEFAULT:0"`
	Title           string `gorm:"index;unique_index:title_datestr"`
	DateStr         string `gorm:"index;unique_index:title_datestr"`
	Summary         string `gorm:"size:2047"`
	Content         string `gorm:"type:mediumtext"`
	PreviewContent  string `gorm:"type:mediumtext"`
	Image           string
	ImageType       string
	AppFeedImg      string // 消息流里经过剪切的缩略图
	PcImage         string
	Url             string
	MediaUrl        string
	Source          string
	DisplayAuthor   string
	LikeCount       int32
	DislikeCount    int32
	PaidCount       int32 `gorm:"default:0;index"`
	Impact          int32 `gorm:"default:0;index"`
	ContentType     sql.NullInt64
	MediaType       sql.NullInt64
	AuthorArticleId int64
	SyncToWscn      bool  `sql:"DEFAULT:0"`
	WscnId          int64 `gorm:"index"` // id returned after being published to wscn live news
	SyncToGlobal    bool  `sql:"DEFAULT:0"`
	GlobalId        int64
	NeedExplained   bool       `sql:"DEFAULT:0" gorm:"index"`
	CrawlerResId    int64      `sql:"DEFAULT:0" gorm:"index"` // id of article in crawler's db
	CrawlerWechatId int64      `sql:"DEFAULT:0" gorm:"index"` // which 公众号 is this article from
	HasXgbXun       bool       `sql:"DEFAULT:1"`
	CreatedAt       time.Time  `gorm:"index"`
	OriginCreatedAt *time.Time // created_at, manual_updated_at 可能被修改，这里存储真实的创建时间。
	// In a query clause like `where created_at > ? and updated_at < ?`,
	// the following index might not be used, because it's a multi range
	// query
	UpdatedAt       time.Time `gorm:"index"`
	ManualUpdatedAt time.Time `gorm:"index"`
	// The following index is used to improved performance of PcNewMsgs()
	// function, caution though it might interfere with index on
	// `created_at` in a query clause like `where deleted_at is NULL and
	// created_at > ?`
	DeletedAt          *time.Time `gorm:"index"`
	IsDeleted          bool       `sql:"DEFAULT:0"`
	PreviewCount       int32      `gorm:"default:0"`
	IsWithdrawn        bool       `gorm:"default:false;index"`
	AILimitationPeriod int        `gorm:"default:0"`       // hours
	UseTempl           bool       `gorm:"index"`           // 是否使用文章模板
	PrettyContent      string     `gorm:"type:mediumtext"` // 阅读模式处理过的文章正文
	PrettyInfo         string     `gorm:"size:1024"`       // 阅读模式内容的相关信息，json 格式
	IsProhibitComments bool
	Watermarks         string `gorm:"size:2047"`
	PrettyIsOriginal   bool   // 阅读模式的内容是否是抓取自有原创标签的文章
	FromFusionOpe      bool

	WhetherHideImpactFace bool `gorm:"default:false"`

	SubTitle string `gorm:"index"`
	Score    int64  // wscn live 同步 1:默认样式 2:红 3:红加粗

	CreatedBy     int64
	IsTodaysFocus bool
}

func (c *ContentMessage) TableName() string {
	return "content_messages"
}
