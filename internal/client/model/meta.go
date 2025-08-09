package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type MetaData struct {
	Key string
	Val string
}

type Meta []MetaData

func NewMeta(data ...string) Meta {
	if len(data)%2 != 0 {
		data = append(data, "")
	}
	meta := make(Meta, len(data)/2)
	for i := range meta {
		i2 := i * 2
		meta[i] = MetaData{
			Key: data[i2],
			Val: data[i2+1],
		}
	}
	return meta
}

func (m Meta) ToString() string {
	d := make([]string, len(m))
	for i := range m {
		d[i] = m[i].Key + ": " + m[i].Val
	}
	return strings.Join(d, ", ")
}

func (m Meta) MarshalWithType(typ Type) ([]byte, error) {
	var buf bytes.Buffer
	if err := buf.WriteByte(byte(typ)); err != nil {
		return nil, fmt.Errorf("failed to write Get byte %v: %w", typ, err)
	}
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return nil, fmt.Errorf("failed to marshal meta: %w", err)
	}
	return buf.Bytes(), nil
}

func UnmarshalMetaAndType(data []byte) (Meta, Type, error) {
	typ := Type(data[0])
	var meta Meta
	if err := json.Unmarshal(data[1:], &meta); err != nil {
		return nil, typ, fmt.Errorf("failed to unmarshal meta: %w", err)
	}
	return meta, typ, nil
}
