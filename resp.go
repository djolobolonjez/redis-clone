package main

import (
	"bufio"
	"io"
	"log"
	"strconv"
)

const (
	STRING  = '+'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
	ERROR   = '-'
	MAP     = '%'
)

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

func (v Value) serializeNull() []byte {
	return []byte("$-1\r\n")
}

func (v Value) serializeBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) serializeString() []byte {
	bytes := make([]byte, 0, 1+len(v.bulk)+2)
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) serializeInteger() []byte {
	var bytes []byte
	bytes = append(bytes, INTEGER)
	bytes = append(bytes, strconv.Itoa(v.num)...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) serializeArray() []byte {
	var bytes []byte
	len := len(v.array)
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.array[i].Serialize()...)
	}

	return bytes
}

func (v Value) serializeError() []byte {
	bytes := make([]byte, 0, 1+len(v.str)+2)
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) serializeMap() []byte {
	return []byte("%0\r\n") // not implemented
}

func (v Value) Serialize() []byte {
	switch v.typ {
	case "bulk":
		return v.serializeBulk()
	case "array":
		return v.serializeArray()
	case "integer":
		return v.serializeInteger()
	case "string":
		return v.serializeString()
	case "null":
		return v.serializeNull()
	case "error":
		return v.serializeError()
	case "map":
		return v.serializeMap()
	default:
		return []byte{}
	}
}

type RespReader struct {
	reader *bufio.Reader
}

func NewRespReader(conn io.Reader) *RespReader {
	return &RespReader{
		reader: bufio.NewReader(conn),
	}
}

type RespWriter struct {
	writer io.Writer
}

func NewRespWriter(conn io.Writer) *RespWriter {
	return &RespWriter{
		writer: conn,
	}
}

func (w *RespWriter) Write(v Value) error {
	bytes := v.Serialize()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (r *RespReader) readLine() (line []byte, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
			return line[:len(line)-2], nil
		}
	}
}

func (r *RespReader) readInteger() (Value, error) {
	v := Value{}
	v.typ = "integer"
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return Value{}, err
	}

	v.num = int(i64)

	return v, nil
}

func (r *RespReader) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"

	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	len, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return Value{}, err
	}

	bulk := make([]byte, len)
	io.ReadFull(r.reader, bulk)

	v.bulk = string(bulk)

	_, err = r.readLine() // flush the input buffer
	if err != nil {
		return Value{}, err
	}

	return v, nil
}

func (r *RespReader) readArray() (Value, error) {
	v := Value{}
	v.typ = "array"

	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	len, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return Value{}, err
	}

	if len == 0 {
		return v, nil
	}

	v.array = make([]Value, len)
	for i := 0; i < int(len); i++ {
		elem, err := r.Read()
		if err != nil {
			return Value{}, err
		}
		v.array[i] = elem
	}

	return v, nil
}

func (r *RespReader) Read() (Value, error) {
	t, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch t {
	case INTEGER:
		return r.readInteger()
	case BULK:
		return r.readBulk()
	case ARRAY:
		return r.readArray()
	default:
		log.Printf("Error: unknown type %v", string(t))
		return Value{}, nil
	}
}
