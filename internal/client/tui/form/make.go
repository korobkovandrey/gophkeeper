package form

import (
	"gophkeeper/internal/client/model"
)

func MakeFormTextModel(id model.ID, meta model.Meta, text string) Model {
	return NewModel(id, newTextModel(text), meta)
}

func MakeFormLoginPassModel(id model.ID, meta model.Meta, login, pass string) Model {
	return NewModel(id, newLoginPassModel(login, pass), meta)
}

func MakeFormCardModel(id model.ID, meta model.Meta, ccn, expire, cvv string) Model {
	return NewModel(id, newCardModel(ccn, expire, cvv), meta)
}
