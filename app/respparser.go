package main

import (
	"bytes"
	"fmt"
)

type RespSimpleString struct {
	Value string
}

func parseSimpleString(data []byte, i int) (res RespSimpleString, end int) {
	for i < len(data)-2 {
		if data[i] == byte('\r') && data[i+1] == byte('\n') {
			break
		}
		res.Value += string(data[i])
		i++
	}

	return res, i + 1
}

type RespError struct {
	Value string
}

type RespInteger struct {
	Value int64
}

func (i RespInteger) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString(":")
	buf.WriteString(fmt.Sprintf("%d", i.Value))
	buf.WriteString("\r\n")
	return buf.Bytes()
}

func ParseInteger(data []byte, i int) (res RespInteger, end int, err error) {
	var sign int64 = 1
	if data[i] == byte('-') {
		sign = -1
		i++
	}
	for i < len(data)-2 {
		if data[i] == byte('\r') && data[i+1] == byte('\n') {
			break
		}
		digit := int64(data[i] - byte('0'))
		if digit < 0 || digit > 9 {
			return RespInteger{}, i, fmt.Errorf("invalid integer digit: %c", data[i])
		}
		res.Value = res.Value*10 + digit
		i++
	}

	res.Value *= sign
	return res, i + 1, nil
}

type RespBulkString struct {
	Value []byte
}

func (bs RespBulkString) Bytes() []byte {
	var buf bytes.Buffer
	if bs.Value == nil {
		buf.WriteString("$-1\r\n")
	} else {
		buf.WriteString("$")
		buf.WriteString(fmt.Sprintf("%d", len(bs.Value)))
		buf.WriteString("\r\n")
		buf.Write(bs.Value)
		buf.WriteString("\r\n")
	}
	return buf.Bytes()
}

func parseBulkString(data []byte, i int) (bs RespBulkString, end int, err error) {
	length, i, err := ParseInteger(data, i)
	if err != nil {
		return RespBulkString{}, i, err
	}
	if length.Value+int64(i) > int64(len(data)) {
		return RespBulkString{}, i, fmt.Errorf("string length out of bounds")
	}
	if length.Value == -1 {
		return RespBulkString{}, i, nil
	}

	end = i + int(length.Value) + 1
	return RespBulkString{data[i+1 : end]}, end + 1, nil
}

func (bs RespBulkString) String() string {
	return string(bs.Value)
}

// interface with Bytes() method to convert to RESP
type Resp interface {
	Bytes() []byte
}

type RespArray struct {
	Value []Resp
}

func (arr *RespArray) Bytes() []byte {
	var buf bytes.Buffer
	if len(arr.Value) == 1 {
		return arr.Value[0].Bytes()
	}

	buf.WriteString("*")
	buf.WriteString(fmt.Sprintf("%d", len(arr.Value)))
	buf.WriteString("\r\n")
	for _, el := range arr.Value {
		buf.Write(el.Bytes())
	}
	return buf.Bytes()
}

func parseArray(data []byte, i int) (arr RespArray, end int, err error) {
	length, i, err := ParseInteger(data, i)
	if err != nil {
		return RespArray{}, i, err
	}
	arr.Value = make([]Resp, length.Value)
	for elc := 0; elc < int(length.Value); elc++ {
		i++
		prefix := data[i]
		i++
		switch prefix {
		case byte('$'):
			bs, end, err := parseBulkString(data, i)
			if err != nil {
				return RespArray{}, i, err
			}
			arr.Value[elc] = bs
			i = end
		default:
			return RespArray{}, i, fmt.Errorf("invalid resp value identifier at position %d: %c", i, prefix)
		}
	}
	return arr, i, err
}
