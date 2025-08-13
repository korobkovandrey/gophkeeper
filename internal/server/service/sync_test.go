package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncService(t *testing.T) {
	s := NewSyncService()
	assert.NotNil(t, s, "SyncService should not be nil")
	assert.NotNil(t, s.subscribers, "subscribers map should be initialized")
	assert.Empty(t, s.subscribers, "subscribers map should be empty")
	assert.False(t, s.closed, "closed should be false")
}

func TestSyncService_Subscribe(t *testing.T) {
	s := NewSyncService()
	userID := int64(1)

	t.Run("Successful subscription", func(t *testing.T) {
		subscriber, err := s.Subscribe(userID)
		assert.NoError(t, err, "expected no error")
		assert.NotNil(t, subscriber, "subscriber should not be nil")
		assert.Equal(t, userID, subscriber.userID, "subscriber userID should match")
		assert.NotNil(t, subscriber.C, "subscriber channel should not be nil")
		assert.Len(t, s.subscribers[userID], 1, "subscribers map should have one entry")
	})

	t.Run("Multiple subscriptions for same user", func(t *testing.T) {
		subscriber1, err1 := s.Subscribe(userID)
		subscriber2, err2 := s.Subscribe(userID)
		assert.NoError(t, err1, "expected no error for first subscription")
		assert.NoError(t, err2, "expected no error for second subscription")
		assert.Len(t, s.subscribers[userID], 3, "subscribers map should have three entries")
		assert.Contains(t, s.subscribers[userID], subscriber1, "subscribers map should contain first subscriber")
		assert.Contains(t, s.subscribers[userID], subscriber2, "subscribers map should contain second subscriber")
	})

	t.Run("Subscribe to closed service", func(t *testing.T) {
		s.Close()
		_, err := s.Subscribe(userID)
		assert.Error(t, err, "expected an error")
		assert.EqualError(t, err, "sync service is closed", "expected closed service error")
	})
}

func TestSyncService_Unsubscribe(t *testing.T) {
	s := NewSyncService()
	userID := int64(1)

	t.Run("Unsubscribe existing subscriber", func(t *testing.T) {
		subscriber, err := s.Subscribe(userID)
		assert.NoError(t, err, "expected no error")
		s.Unsubscribe(subscriber)
		assert.Len(t, s.subscribers[userID], 0, "subscribers map should be empty")
		select {
		case _, open := <-subscriber.C:
			assert.False(t, open, "subscriber channel should be closed")
		default:
			t.Error("expected subscriber channel to be closed")
		}
	})

	t.Run("Unsubscribe non-existing subscriber", func(t *testing.T) {
		subscriber := &Subscriber{userID: userID, C: make(chan SyncMsg, 10)}
		s.Unsubscribe(subscriber)
		assert.Empty(t, s.subscribers, "subscribers map should remain empty")
		select {
		case <-subscriber.C:
			t.Error("expected subscriber channel not have any message and not to be closed")
		default:
		}
	})
}

func TestSyncService_Broadcast(t *testing.T) {
	s := NewSyncService()
	userID := int64(1)
	secret := Secret{
		ID:        "secret-123",
		UserID:    userID,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("metadata"),
		Data:      []byte("secret-data"),
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}
	syncMsg := SyncMsg{
		Action: 1,
		ID:     "secret-123",
		Secret: secret,
	}

	t.Run("Broadcast to single subscriber", func(t *testing.T) {
		subscriber, err := s.Subscribe(userID)
		assert.NoError(t, err, "expected no error")

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := <-subscriber.C
			assert.Equal(t, syncMsg, msg, "received message should match sent message")
		}()

		s.Broadcast(syncMsg.Action, syncMsg.ID, syncMsg.Secret)
		wg.Wait()
	})

	t.Run("Broadcast to multiple subscribers", func(t *testing.T) {
		subscriber1, err1 := s.Subscribe(userID)
		assert.NoError(t, err1, "expected no error for first subscription")
		subscriber2, err2 := s.Subscribe(userID)
		assert.NoError(t, err2, "expected no error for second subscription")

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			msg := <-subscriber1.C
			assert.Equal(t, syncMsg, msg, "received message should match sent message for subscriber1")
		}()
		go func() {
			defer wg.Done()
			msg := <-subscriber2.C
			assert.Equal(t, syncMsg, msg, "received message should match sent message for subscriber2")
		}()

		s.Broadcast(syncMsg.Action, syncMsg.ID, syncMsg.Secret)
		wg.Wait()
	})

	t.Run("Broadcast to non-existing user", func(t *testing.T) {
		s.Broadcast(syncMsg.Action, syncMsg.ID, syncMsg.Secret)
		// No subscribers, so no assertions needed; just verify no panic
	})
}

func TestSyncService_Close(t *testing.T) {
	s := NewSyncService()
	userID1 := int64(1)
	userID2 := int64(2)

	subscriber1, err1 := s.Subscribe(userID1)
	assert.NoError(t, err1, "expected no error for userID1 subscription")
	subscriber2, err2 := s.Subscribe(userID2)
	assert.NoError(t, err2, "expected no error for userID2 subscription")

	s.Close()

	assert.True(t, s.closed, "service should be closed")
	assert.Empty(t, s.subscribers, "subscribers map should be empty")
	_, open1 := <-subscriber1.C
	assert.False(t, open1, "subscriber1 channel should be closed")
	_, open2 := <-subscriber2.C
	assert.False(t, open2, "subscriber2 channel should be closed")

	// Verify new subscriptions are rejected
	_, err := s.Subscribe(userID1)
	assert.Error(t, err, "expected an error")
	assert.EqualError(t, err, "sync service is closed", "expected closed service error")
}
