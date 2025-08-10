package service

import "time"

type Time struct {
	diff time.Duration
}

func NewTime() *Time {
	return &Time{}
}

func (s *Time) SetDiff(diff time.Duration) {
	s.diff = diff.Truncate(time.Second)
}

func (s *Time) ServerNow() time.Time {
	return s.Server(time.Now().Truncate(time.Second))
}

func (s *Time) Server(local time.Time) time.Time {
	return local.Add(s.diff)
}

func (s *Time) Local(server time.Time) time.Time {
	return server.Add(-s.diff)
}

func (s *Time) LocalFromUnix(sec int64) time.Time {
	return s.Local(time.Unix(sec, 0))
}
