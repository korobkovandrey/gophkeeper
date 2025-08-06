package tuiadapter

type DataType int

const (
	DataTypeUnknown DataType = iota
	DataTypeText
)

type MetaData struct {
	Key string
	Val string
}
type Meta []MetaData

func NewMeta(data ...string) Meta {
	if len(data)%2 != 0 {
		data = append(data, "")
	}
	meta := make(Meta, len(data)/2)
	for i := range meta {
		i2 := i * 2
		meta[i] = MetaData{
			Key: data[i2],
			Val: data[i2+1],
		}
	}
	return meta
}

type DataText struct {
	ID   string
	Meta Meta
	Text string
}
