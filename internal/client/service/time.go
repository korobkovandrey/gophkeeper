package service

import "time"

type Time struct {
	Diff time.Duration
}

func NewTime() *Time {
	return &Time{}
}

func (s *Time) SetDiff(diff time.Duration) {
	s.Diff = diff
}

func (s *Time) ServerNow() time.Time {
	return s.Server(time.Now())
}

func (s *Time) Server(local time.Time) time.Time {
	return local.Add(s.Diff)
}

func (s *Time) Local(server time.Time) time.Time {
	return server.Add(-s.Diff)
}
