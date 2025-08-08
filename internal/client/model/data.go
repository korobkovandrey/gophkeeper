package model

// DataText текстовые данные
type DataText struct {
	Text string
}

// DataLoginPass логин/пароль
type DataLoginPass struct {
	Login string
	Pass  string
}

// DataCard данные карты
type DataCard struct {
	Number string
	Expire string
	CCN    string
}
