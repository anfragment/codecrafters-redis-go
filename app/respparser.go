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

func parseInteger(data []byte, i int) (res RespInteger, end int, err error) {
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

	return res, i + 1, nil
}

type RespBulkString struct {
	Value []byte
}

func parseBulkString(data []byte, i int) (bs RespBulkString, end int, err error) {
	length, i, err := parseInteger(data, i)
	if err != nil {
		return RespBulkString{}, i, err
	}
	if length.Value+int64(i) > int64(len(data)) {
		return RespBulkString{}, i, fmt.Errorf("string length out of bounds")
	}

	end = i + int(length.Value) + 1
	return RespBulkString{data[i+1 : end]}, end + 1, nil
}

func (bs RespBulkString) String() string {
	return string(bs.Value)
}

type RespArray struct {
	Value []interface{}
}

func (arr *RespArray) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString("*")
	buf.WriteString(fmt.Sprintf("%d", len(arr.Value)))
	buf.WriteString("\r\n")
	for _, el := range arr.Value {
		switch v := el.(type) {
		case RespBulkString:
			buf.WriteString("$")
			buf.WriteByte(byte(len(v.Value)) + byte('0'))
			buf.WriteString("\r\n")
			buf.Write(v.Value)
			buf.WriteString("\r\n")
		case RespInteger:
			buf.WriteString(":")
			buf.WriteString(fmt.Sprintf("%d", v.Value))
			buf.WriteString("\r\n")
		}
	}
	return buf.Bytes()
}

func parseArray(data []byte, i int) (arr RespArray, end int, err error) {
	length, i, err := parseInteger(data, i)
	if err != nil {
		return RespArray{}, i, err
	}
	arr.Value = make([]interface{}, length.Value)
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
