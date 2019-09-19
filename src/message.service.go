package src

import (
	"sync"

	"github.com/jinzhu/gorm"
)

type MessageService struct {
	db           *gorm.DB
	data         []Message
	mutex        *sync.Mutex
	currentIndex int
	dataLength   int
}

func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{
		db:           db,
		data:         make([]Message, 100),
		mutex:        &sync.Mutex{},
		currentIndex: 100,
		dataLength:   0,
	}
}

func (s *MessageService) Create(m Message) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.dataLength++
	s.currentIndex--
	if s.currentIndex < 0 {
		s.currentIndex = len(s.data) - 1
	}
	s.data[s.currentIndex] = m
	return nil
}

func (s *MessageService) GetList() (ms []Message, err error) {
	i := s.currentIndex
	if s.dataLength < 100 {
		ms = s.data[i:100]
	} else {
		ms = append(s.data[i:100], s.data[0:i]...)[0:100]
	}
	return
}
