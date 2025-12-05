package dao

import (
	"time"
)

type ActivityFreeSubject struct {
	ID              int64 `gorm:"primarykey"`
	UserId          int
	SubjectId       int
	SubjectFreeDays int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a *ActivityFreeSubject) TableName() string {
	return "activity_free_subjects"
}
