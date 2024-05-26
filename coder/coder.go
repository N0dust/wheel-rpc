package coder

import "io"

type Header struct {
	ServiceMethod string // service.method
	Seq           uint64
	Error         string
}

type Coder interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewSerializer func(closer io.ReadWriteCloser) Coder

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCoderFuncMap map[Type]NewSerializer

func init() {
	NewCoderFuncMap = make(map[Type]NewSerializer)
	NewCoderFuncMap[GobType] = NewGobCoder
}
