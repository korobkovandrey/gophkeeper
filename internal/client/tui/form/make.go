package form

import (
	"gophkeeper/internal/client/model"
)

func MakeFormTextModel(id model.ID, meta model.Meta, text string) Model {
	return NewModel(id, newTextModel(text), meta)
}
