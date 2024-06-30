package resp

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

var (
	ErrInvalidSyntax = errors.New("invalid syntax")
)

const (
	prefixSimpleString = '+'
	prefixError        = '-'
	prefixInteger      = ':'
	prefixBulkString   = '$'
	prefixArray        = '*'
)

var (
	ErrMarshal = errors.New("failed to marshal value")
)

type ReaderPeeker interface {
	io.Reader
	Peek(n int) ([]byte, error)
}

func Unmarshal(r ReaderPeeker) (interface{}, error) {
	return unmarshalAny(r)
}

func Marshal(w io.Writer, value interface{}) error {
	return marshalAny(w, value)
}

func marshalArray(w io.Writer, value []interface{}) error {
	var err error
	if value == nil {
		_, err = fmt.Fprintf(w, "*-1\r\n")
	} else {
		_, err = fmt.Fprintf(w, "*%d\r\n", len(value))
	}

	if err != nil {
		return err
	}
	for _, item := range value {
		if err = marshalAny(w, item); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalArray(r ReaderPeeker) ([]interface{}, error) {
	rawHeader, err := readUntilCRLF(r)
	if err != nil {
		return nil, err
	}

	size, err := strconv.ParseInt(string(rawHeader[1:]), 10, 64)
	if err != nil {
		return nil, err
	}

	if size < -1 {
		return nil, fmt.Errorf("%w: size of an array must not be less than -1", ErrInvalidSyntax)
	}

	if size == -1 {
		return nil, nil
	}

	value := make([]interface{}, size)
	for i := int64(0); i < size; i++ {
		if value[i], err = unmarshalAny(r); err != nil {
			return nil, err
		}
	}

	return value, nil
}

func marshalSimpleString(w io.Writer, value string) error {
	if strings.Index(value, "\r\n") != -1 {
		return fmt.Errorf("%w: simple string contains CRLF", ErrMarshal)
	}
	_, err := fmt.Fprintf(w, "+%s\r\n", value)
	return err
}

func unmarshalSimpleString(r ReaderPeeker) (string, error) {
	data, err := readUntilCRLF(r)
	if err != nil {
		return "", err
	}
	if len(data) < 1 || data[0] != prefixSimpleString {
		return "", fmt.Errorf("%w: simple string must start with '+'", ErrInvalidSyntax)
	}
	return string(data[1:]), nil
}

type Integer interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

func marshalInt[T Integer](w io.Writer, value T) error {
	_, err := fmt.Fprintf(w, ":%d\r\n", value)
	return err
}

func unmarshalInt[T Integer](r ReaderPeeker) (T, error) {
	var value T
	data, err := readUntilCRLF(r)
	if err != nil {
		return value, err
	}
	if len(data) < 1 || data[0] != prefixInteger {
		return value, fmt.Errorf("%w: simple string must start with '+'", ErrInvalidSyntax)
	}

	strInt := string(data[1:])
	bitSize := int(unsafe.Sizeof(value)) * 8
	switch any(value).(type) {
	case int, int64, int32, int16, int8:
		var i int64
		i, err = strconv.ParseInt(strInt, 10, bitSize)
		return T(i), nil
	case uint, uint64, uint32, uint16, uint8:
		var i uint64
		i, err = strconv.ParseUint(strInt, 10, bitSize)
		return T(i), nil
	default:
		panic("never")
	}
}

func marshalError(w io.Writer, value error) error {
	data := value.Error()
	if strings.Index(data, "\r\n") != -1 {
		return fmt.Errorf("%w: error contains CRLF", ErrMarshal)
	}
	_, err := fmt.Fprintf(w, "-%s\r\n", data)
	return err
}

func unmarshalError(r ReaderPeeker) (error, error) {
	data, err := readUntilCRLF(r)
	if err != nil {
		return nil, err
	}
	if len(data) < 1 || data[0] != prefixError {
		return nil, fmt.Errorf("%w: error string must start with '-'", ErrInvalidSyntax)
	}
	return errors.New(string(data[1:])), nil
}

func marshalBulkString(w io.Writer, val []byte) error {
	if val == nil {
		_, err := fmt.Fprintf(w, "$-1\r\n")
		return err
	}
	_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(val), val)

	return err
}

func unmarshalBulkString(r ReaderPeeker) ([]byte, error) {
	rawHeader, err := readUntilCRLF(r)
	if err != nil {
		return nil, err
	}
	if rawHeader[0] != prefixBulkString {
		return nil, fmt.Errorf("%w: bulk string must start with '$'", ErrInvalidSyntax)
	}

	size, err := strconv.ParseInt(string(rawHeader[1:]), 10, 64)
	if err != nil {
		return nil, err
	}

	if size == -1 {
		return nil, nil
	}

	if size == 0 {
		return []byte{}, nil
	}

	rawData := make([]byte, size+2)
	if n, err := r.Read(rawData); err != nil || n != len(rawData) {
		return nil, fmt.Errorf("%w: can't read enoguh data", ErrInvalidSyntax)
	}

	if string(rawData[len(rawData)-2:]) != "\r\n" {
		return nil, fmt.Errorf("%w: not null bulk string must has crlf ending", ErrInvalidSyntax)
	}
	return rawData[:len(rawData)-2], nil
}

func unmarshalAny(r ReaderPeeker) (interface{}, error) {
	prefix, err := r.Peek(1)
	if err != nil {
		return nil, err
	}
	switch prefix[0] {
	case prefixSimpleString:
		return unmarshalSimpleString(r)
	case prefixArray:
		return unmarshalArray(r)
	case prefixInteger:
		return unmarshalInt[int64](r)
	case prefixError:
		return unmarshalError(r)
	case prefixBulkString:
		return unmarshalBulkString(r)
	default:
		return nil, fmt.Errorf("%w: unknown prefix %s", ErrMarshal, prefix)
	}
}

func marshalAny(w io.Writer, value interface{}) error {
	switch v := value.(type) {
	case int:
		return marshalInt[int](w, v)
	case int64:
		return marshalInt[int64](w, v)
	case uint64:
		return marshalInt[uint64](w, v)
	case uint32:
		return marshalInt[uint32](w, v)
	case int32:
		return marshalInt[int32](w, v)
	case string:
		return marshalSimpleString(w, v)
	case []byte:
		return marshalBulkString(w, v)
	case error:
		return marshalError(w, v)
	case []interface{}:
		return marshalArray(w, v)
	default:
		panic("unsupported type " + reflect.TypeOf(v).String())
	}
}

func readUntilCRLF(r ReaderPeeker) ([]byte, error) {
	result := make([]byte, 0, 32)
	buf := make([]byte, 1)
	for {
		if _, err := r.Read(buf); err != nil {
			return nil, err
		}
		if buf[0] == '\n' && result[len(result)-1] == '\r' {
			return result[:len(result)-1], nil
		}
		result = append(result, buf[0])
	}
}
