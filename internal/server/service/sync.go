package service

import (
	"errors"
	"sync"
)

type SyncMsg struct {
	Action int32
	ID     string
	Secret Secret
}

type Subscriber struct {
	userID int64
	C      chan SyncMsg
}

type SyncService struct {
	mux         sync.Mutex
	subscribers map[int64]map[*Subscriber]*Subscriber
	closed      bool
}

func NewSyncService() *SyncService {
	return &SyncService{
		subscribers: make(map[int64]map[*Subscriber]*Subscriber),
	}
}

func (s *SyncService) Subscribe(userID int64) (*Subscriber, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.closed {
		return nil, errors.New("sync service is closed")
	}
	if _, ok := s.subscribers[userID]; !ok {
		s.subscribers[userID] = make(map[*Subscriber]*Subscriber)
	}
	subscriber := &Subscriber{
		userID: userID,
		C:      make(chan SyncMsg, 10),
	}
	s.subscribers[userID][subscriber] = subscriber
	return subscriber, nil
}

func (s *SyncService) Unsubscribe(subscriber *Subscriber) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.subscribers[subscriber.userID]; !ok {
		return
	}
	delete(s.subscribers[subscriber.userID], subscriber)
	close(subscriber.C)
}

func (s *SyncService) Broadcast(action int32, id string, secret Secret) {
	s.mux.Lock()
	defer s.mux.Unlock()
	subscribers, ok := s.subscribers[secret.UserID]
	if !ok {
		return
	}
	syncMsg := SyncMsg{
		ID:     id,
		Action: action,
		Secret: secret,
	}
	for _, subscriber := range subscribers {
		subscriber.C <- syncMsg
	}
}

func (s *SyncService) Close() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.closed = true
	for _, subscribers := range s.subscribers {
		for _, subscriber := range subscribers {
			delete(s.subscribers[subscriber.userID], subscriber)
			close(subscriber.C)
		}
	}
	s.subscribers = make(map[int64]map[*Subscriber]*Subscriber)
}
