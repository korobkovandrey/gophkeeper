package form

import "gophkeeper/internal/client/tuiadapter"

func MakeFormTextModel(id, text string, meta tuiadapter.Meta) Model {
	return NewModel(id, newTextModel(text), meta)
}
