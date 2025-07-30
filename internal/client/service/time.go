package service

import "time"

type Time struct {
	diff time.Duration
}

func NewTime() *Time {
	return &Time{}
}

func (s *Time) SetDiff(diff time.Duration) {
	s.diff = diff
}

func (s *Time) Current() time.Time {
	return time.Now().Add(s.diff)
}
