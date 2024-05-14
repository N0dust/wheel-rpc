package serializer

import "io"

type Header struct {
	ServiceMethod string // service.method
	Seq           uint64
	Error         string
}

type Serialize interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewSerialize func(closer io.ReadWriteCloser) Serialize

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewSerializeFuncMap map[Type]NewSerialize

func init() {
	NewSerializeFuncMap = make(map[Type]NewSerialize)
	NewSerializeFuncMap[GobType] = NewGobSerializer
}
