package serializer

import "io"

type Header struct {
	ServiceMethod string // service.method
	Seq           uint64
	Error         string
}

type Serializer interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewSerializer func(closer io.ReadWriteCloser) Serializer

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewSerializerFuncMap map[Type]NewSerializer

func init() {
	NewSerializerFuncMap = make(map[Type]NewSerializer)
	NewSerializerFuncMap[GobType] = NewGobSerializer
}
