package coder

import (
	"bufio"
	"encoding/gob"
	"io"
)

type GobCoder struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func (s *GobCoder) Close() error {
	return s.conn.Close()
}

func (s *GobCoder) ReadHeader(header *Header) error {
	return s.dec.Decode(header)
}

func (s *GobCoder) ReadBody(i interface{}) error {
	return s.dec.Decode(i)
}

func (s *GobCoder) Write(header *Header, i interface{}) error {
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

func NewGobCoder(conn io.ReadWriteCloser) Coder {
	buf := bufio.NewWriter(conn)
	return &GobCoder{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(conn),
	}
}
