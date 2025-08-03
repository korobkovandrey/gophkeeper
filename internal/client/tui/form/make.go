package form

func MakeFormTextModel(id, text string, metaKeys, metaVals []string) Model {
	return NewModel(id, newTextModel(text), metaKeys, metaVals)
}
