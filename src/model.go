package src

import "time"

type Message struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Message   string `json:"msg" gorm:"column:message;type:text"`
	Timestamp int64  `json:"ts" gorm:"column:timestamp;index"`
}
