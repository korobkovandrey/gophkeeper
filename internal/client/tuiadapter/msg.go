package tuiadapter

type ChangeMsg struct {
	ID string
}

func NewChangeMsg(id string) ChangeMsg {
	return ChangeMsg{ID: id}
}
