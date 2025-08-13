package model

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected Meta
	}{
		{
			name:     "Even number of inputs",
			input:    []string{"key1", "val1", "key2", "val2"},
			expected: Meta{{Key: "key1", Val: "val1"}, {Key: "key2", Val: "val2"}},
		},
		{
			name:     "Odd number of inputs",
			input:    []string{"key1", "val1", "key2"},
			expected: Meta{{Key: "key1", Val: "val1"}, {Key: "key2", Val: ""}},
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: Meta{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMeta(tt.input...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetaToString(t *testing.T) {
	tests := []struct {
		name     string
		meta     Meta
		expected string
	}{
		{
			name:     "Multiple key-value pairs",
			meta:     Meta{{Key: "key1", Val: "val1"}, {Key: "key2", Val: "val2"}},
			expected: "key1: val1, key2: val2",
		},
		{
			name:     "Single key-value pair",
			meta:     Meta{{Key: "key1", Val: "val1"}},
			expected: "key1: val1",
		},
		{
			name:     "Empty meta",
			meta:     Meta{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.meta.ToString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetaMarshalWithType(t *testing.T) {
	tests := []struct {
		name     string
		meta     Meta
		typ      Type
		expected func([]byte) bool
	}{
		{
			name: "Valid meta with TypeText",
			meta: Meta{{Key: "key1", Val: "val1"}},
			typ:  TypeText,
			expected: func(b []byte) bool {
				return b[0] == byte(TypeText) && len(b) > 1
			},
		},
		{
			name: "Empty meta with TypeCard",
			meta: Meta{},
			typ:  TypeCard,
			expected: func(b []byte) bool {
				return b[0] == byte(TypeCard) && len(b) > 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.meta.MarshalWithType(tt.typ)
			assert.NoError(t, err)
			assert.True(t, tt.expected(result), "Unexpected marshaled output")
		})
	}
}

func TestUnmarshalMetaAndType(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		expectedMeta Meta
		expectedType Type
		expectError  bool
	}{
		{
			name: "Valid input with TypeText",
			input: func() []byte {
				var buf bytes.Buffer
				buf.WriteByte(byte(TypeText))
				_ = json.NewEncoder(&buf).Encode(Meta{{Key: "key1", Val: "val1"}})
				return buf.Bytes()
			}(),
			expectedMeta: Meta{{Key: "key1", Val: "val1"}},
			expectedType: TypeText,
			expectError:  false,
		},
		{
			name:         "Invalid JSON",
			input:        []byte{byte(TypeLoginPass), '{', '}'},
			expectedMeta: nil,
			expectedType: TypeLoginPass,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, typ, err := UnmarshalMetaAndType(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMeta, meta)
				assert.Equal(t, tt.expectedType, typ)
			}
		})
	}
}
