package serializer

import (
	"bufio"
	"encoding/gob"
	"io"
)

type GobSerializer struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func (s *GobSerializer) Close() error {
	return s.conn.Close()
}

func (s *GobSerializer) ReadHeader(header *Header) error {
	return s.dec.Decode(header)
}

func (s *GobSerializer) ReadBody(i interface{}) error {
	return s.dec.Decode(i)
}

func (s *GobSerializer) Write(header *Header, i interface{}) error {
	var err error
	defer func() {
		_ = s.buf.Flush()
		if err != nil {
			_ = s.Close()
		}
	}()

	err = s.enc.Encode(header)
	if err != nil {
		return err
	}
	err = s.enc.Encode(i)
	if err != nil {
		return err
	}

	return err
}

func NewGobSerializer(conn io.ReadWriteCloser) Serialize {
	buf := bufio.NewWriter(conn)
	return &GobSerializer{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(conn),
	}
}
