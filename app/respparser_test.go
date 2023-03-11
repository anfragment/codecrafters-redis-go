package main

import (
	"bytes"
	"testing"
)

func TestParseSimpleString(t *testing.T) {
	str, i := parseSimpleString([]byte("OK\r\n"), 0)
	if str.Value != "OK" || i != 3 {
		t.Errorf("parseSimpleString([]byte(\"OK\\r\\n\")) = %s, %d, want OK, 4", str.Value, i)
	}
}

func TestParseInteger(t *testing.T) {
	i, end, err := parseInteger([]byte("123456\r\n"), 0)
	if i.Value != 123456 || end != 7 || err != nil {
		t.Errorf(`parseInteger([]byte("123456\r\n") = %v, %d, %v, want 123456, 7, nil`, i, end, err)
	}
}

func TestParseBulkString(t *testing.T) {
	bs, end, err := parseBulkString([]byte("5\r\nhello\r\n"), 0)
	if !bytes.Equal(bs.Value, []byte("hello")) || end != 9 || err != nil {
		t.Errorf(`parseBulkString([]byte("$5\r\nhello\r\n"), 0) = %v, %d, %v, want RespBulkString{Value:[]byte("hello")}, 9, nil`, bs, end, err)
	}
}

func TestParseArray(t *testing.T) {
	arr, end, err := parseArray([]byte("2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n"), 0)
	if len(arr.Value) != 2 || end != 24 || err != nil {
		t.Errorf(`parseArray([]byte("2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n"), 0) = %v, %d, %v, want RespArray{Value:[]Resp{RespBulkString{Value:[]byte("LLEN")}, RespBulkString{Value:[]byte("mylist")}}}, 24, nil`, arr, end, err)
	}
}
