package form

import (
	"gophkeeper/internal/client/model"

	"github.com/charmbracelet/bubbles/textinput"
)

func addMetas(m *Model, meta model.Meta) {
	for i := range meta {
		key := textinput.New()
		key.Width = 10
		key.Placeholder = "Key"
		key.Prompt = prompt
		key.SetValue(meta[i].Key)
		m.metaKeys = append(m.metaKeys, key)
		val := textinput.New()
		val.Width = 10
		val.Placeholder = "Value"
		val.Prompt = prompt
		val.SetValue(meta[i].Val)
		m.metaVals = append(m.metaVals, val)
	}
}

func (m Model) meta() model.Meta {
	var meta model.Meta
	for i := range m.metaKeys {
		k := m.metaKeys[i].Value()
		v := m.metaVals[i].Value()
		if k == "" && v == "" {
			continue
		}
		meta = append(meta, model.MetaData{
			Key: k,
			Val: v,
		})
	}
	return meta
}
