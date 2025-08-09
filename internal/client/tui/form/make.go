package form

import (
	"gophkeeper/internal/client/model"
)

func MakeFormTextModel(id, newID model.ID, meta model.Meta, text string) Model {
	return NewModel(id, newID, newTextModel(text), meta)
}

func MakeFormLoginPassModel(id, newID model.ID, meta model.Meta, login, pass string) Model {
	return NewModel(id, newID, newLoginPassModel(login, pass), meta)
}

func MakeFormCardModel(id, newID model.ID, meta model.Meta, ccn, expire, cvv string) Model {
	return NewModel(id, newID, newCardModel(ccn, expire, cvv), meta)
}
