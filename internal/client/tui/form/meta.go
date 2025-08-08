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
		key.Prompt = "> "
		key.SetValue(meta[i].Key)
		m.metaKeys = append(m.metaKeys, key)
		val := textinput.New()
		val.Width = 10
		val.Placeholder = "Value"
		val.Prompt = "> "
		val.SetValue(meta[i].Val)
		m.metaVals = append(m.metaVals, val)
	}
}

func (m Model) meta() model.Meta {
	meta := make(model.Meta, len(m.metaKeys))
	for i := range m.metaKeys {
		meta[i] = model.MetaData{
			Key: m.metaKeys[i].Value(),
			Val: m.metaVals[i].Value(),
		}
	}
	return meta
}
