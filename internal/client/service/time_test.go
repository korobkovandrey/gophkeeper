package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTime(t *testing.T) {
	tm := NewTime()
	assert.NotNil(t, tm, "NewTime should return a non-nil Time")
	assert.Equal(t, time.Duration(0), tm.diff, "Diff should be zero")
}

func TestSetDiff(t *testing.T) {
	tm := NewTime()
	tests := []struct {
		name     string
		input    time.Duration
		expected time.Duration
	}{
		{
			name:     "Positive duration",
			input:    5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "Negative duration",
			input:    -5 * time.Second,
			expected: -5 * time.Second,
		},
		{
			name:     "Duration with milliseconds",
			input:    5*time.Second + 500*time.Millisecond,
			expected: 5 * time.Second, // Truncated to seconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetDiff(tt.input)
			assert.Equal(t, tt.expected, tm.diff, "Diff should be set correctly")
		})
	}
}

func TestServerNow(t *testing.T) {
	tm := NewTime()
	tm.SetDiff(2 * time.Second)

	now := time.Now().Truncate(time.Second)
	result := tm.ServerNow()

	assert.Equal(t, now.Add(2*time.Second), result, "ServerNow should add diff to current time")
}

func TestServer(t *testing.T) {
	tm := NewTime()
	tests := []struct {
		name     string
		diff     time.Duration
		local    time.Time
		expected time.Time
	}{
		{
			name:     "Positive diff",
			diff:     2 * time.Second,
			local:    time.Unix(1000, 0),
			expected: time.Unix(1002, 0),
		},
		{
			name:     "Negative diff",
			diff:     -3 * time.Second,
			local:    time.Unix(1000, 0),
			expected: time.Unix(997, 0),
		},
		{
			name:     "Zero diff",
			diff:     0,
			local:    time.Unix(1000, 0),
			expected: time.Unix(1000, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetDiff(tt.diff)
			result := tm.Server(tt.local)
			assert.Equal(t, tt.expected, result, "Server should adjust time by diff")
		})
	}
}

func TestLocal(t *testing.T) {
	tm := NewTime()
	tests := []struct {
		name     string
		diff     time.Duration
		server   time.Time
		expected time.Time
	}{
		{
			name:     "Positive diff",
			diff:     2 * time.Second,
			server:   time.Unix(1002, 0),
			expected: time.Unix(1000, 0),
		},
		{
			name:     "Negative diff",
			diff:     -3 * time.Second,
			server:   time.Unix(997, 0),
			expected: time.Unix(1000, 0),
		},
		{
			name:     "Zero diff",
			diff:     0,
			server:   time.Unix(1000, 0),
			expected: time.Unix(1000, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetDiff(tt.diff)
			result := tm.Local(tt.server)
			assert.Equal(t, tt.expected, result, "Local should adjust time by negative diff")
		})
	}
}

func TestLocalFromUnix(t *testing.T) {
	tm := NewTime()
	tests := []struct {
		name     string
		diff     time.Duration
		unixSec  int64
		expected time.Time
	}{
		{
			name:     "Positive diff",
			diff:     2 * time.Second,
			unixSec:  1002,
			expected: time.Unix(1000, 0),
		},
		{
			name:     "Negative diff",
			diff:     -3 * time.Second,
			unixSec:  997,
			expected: time.Unix(1000, 0),
		},
		{
			name:     "Zero diff",
			diff:     0,
			unixSec:  1000,
			expected: time.Unix(1000, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetDiff(tt.diff)
			result := tm.LocalFromUnix(tt.unixSec)
			assert.Equal(t, tt.expected, result, "LocalFromUnix should adjust unix time by negative diff")
		})
	}
}
