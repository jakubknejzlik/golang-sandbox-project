package src

import (
	"github.com/jinzhu/gorm"
)

type MessageService struct {
	db *gorm.DB
}

func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db}
}

func (s *MessageService) Create(m *Message) error {
	return s.db.Create(m).Error
}

func (s *MessageService) GetList(limit int) (ms []Message, err error) {
	err = s.db.Find(&ms).Limit(limit).Order("ts", true).Error
	return
}
