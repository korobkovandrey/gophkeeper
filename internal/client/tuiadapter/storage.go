package tuiadapter

import (
	"strconv"
	"time"
)

type Storage struct {
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Type(id string) (DataType, error) {
	return DataTypeText, nil
}

type Row struct {
	Status  string
	Type    string
	ID      string
	Meta    string
	Created string
	Updated string
}

func (s *Storage) List() []Row {
	rows := make([]Row, 5)
	for i := range rows {
		rows[i] = Row{
			Status:  "Сохранен",
			Type:    "Текст",
			ID:      "id" + strconv.Itoa(i),
			Meta:    "meta",
			Created: time.Now().Add(-time.Duration(i-1) * time.Minute).Format(time.DateTime),
			Updated: time.Now().Add(-time.Duration(i) * time.Minute).Format(time.DateTime),
		}
	}
	return rows
}

func (s *Storage) DataText(id string) (DataText, error) {
	return DataText{
		ID:   id,
		Meta: NewMeta("ss", "ff"),
		Text: "sadsadada",
	}, nil
}

func (s *Storage) SaveText(id, newID, text string, meta Meta) error {
	return nil
}

func (s *Storage) Delete(id string) error {
	return nil
}
