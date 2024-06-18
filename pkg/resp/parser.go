package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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

type ValueType string

var (
	ErrMarshal = errors.New("failed to marshal value")
)

const (
	TypeInteger      ValueType = "i"
	TypeBulkString   ValueType = "bs"
	TypeSimpleString ValueType = "s"
	TypeArray        ValueType = "a"
	TypeError        ValueType = "e"
)

type ReaderPeeker interface {
	io.Reader
	Peek(n int) ([]byte, error)
}

type Marshaller interface {
	Marshal(w io.Writer) error
	Unmarshal(r ReaderPeeker) error
}

type Value interface {
	Marshaller
	Type() ValueType
	Value() interface{}
	IsNull() bool
	String() (val string, ok bool)
	Int() (val int64, ok bool)
	Array() (val []Value, ok bool)
	Bytes() (val []byte, ok bool)
	Error() (val string, ok bool)
}

type baseValue[T any] struct {
	value T
}

func (b *baseValue[T]) Value() interface{} {
	return b.value
}

func (b *baseValue[T]) Type() ValueType {
	panic("not implemented")
}

func (b *baseValue[T]) IsNull() bool {
	return false
}

func (b *baseValue[T]) Error() (val string, ok bool) {
	return "", false
}

func (b *baseValue[T]) String() (val string, ok bool) {
	return "", false
}

func (b *baseValue[T]) Int() (val int64, ok bool) {
	return 0, false
}

func (b *baseValue[T]) Array() (val []Value, ok bool) {
	return nil, false
}

func (b *baseValue[T]) Bytes() (val []byte, ok bool) {
	return nil, false
}

type array struct {
	baseValue[[]Value]
	isNull bool
}

func Array(values []Value) Value {
	return &array{
		baseValue: baseValue[[]Value]{value: values},
		isNull:    false,
	}
}

func NullArray() Value {
	return &array{
		baseValue: baseValue[[]Value]{value: nil},
		isNull:    false,
	}
}

func (a *array) Type() ValueType {
	return TypeArray
}

func (a *array) Array() (val []Value, ok bool) {
	return a.value, true
}

func (a *array) IsNull() bool {
	return a.isNull
}

func (a *array) Marshal(w io.Writer) error {
	_, err := fmt.Fprintf(w, "*%d\r\n", len(a.value))
	if err != nil {
		return err
	}
	for _, item := range a.value {
		if err = item.Marshal(w); err != nil {
			return err
		}
	}
	return nil
}

func (a *array) Unmarshal(r ReaderPeeker) error {
	rawHeader, err := readUntilCRLF(r)
	if err != nil {
		return err
	}

	size, err := strconv.ParseInt(string(rawHeader[1:]), 10, 64)
	if err != nil {
		return err
	}

	if size < -1 {
		return fmt.Errorf("%w: size of an array must not be less than -1", ErrInvalidSyntax)
	}

	if size == -1 {
		a.value = nil
		a.isNull = true
		return nil
	}

	a.value = make([]Value, size)
	for i := int64(0); i < size; i++ {
		if a.value[i], err = unmarshalAny(r); err != nil {
			return err
		}
	}

	return nil
}

type simpleString struct {
	baseValue[string]
}

func SimpleString(value string) Value {
	return &simpleString{baseValue[string]{value: value}}
}

func (s *simpleString) Type() ValueType {
	return TypeSimpleString
}

func (s *simpleString) String() (val string, ok bool) {
	return s.value, true
}

func (s *simpleString) Marshal(w io.Writer) error {
	if strings.Index(s.value, "\r\n") != -1 {
		return fmt.Errorf("%w: simple string contains CRLF", ErrMarshal)
	}
	_, err := fmt.Fprintf(w, "+%s\r\n", s.value)
	return err
}

func (s *simpleString) Unmarshal(r ReaderPeeker) error {
	data, err := readUntilCRLF(r)
	if err != nil {
		return err
	}
	if len(data) < 1 || data[0] != prefixSimpleString {
		return fmt.Errorf("%w: simple string must start with '+'", ErrInvalidSyntax)
	}
	s.value = string(data[1:])
	return nil
}

type integer struct {
	baseValue[int64]
}

func Int(value int64) Value {
	return &integer{baseValue[int64]{value: value}}
}

func (i *integer) Type() ValueType {
	return TypeInteger
}

func (i *integer) Int() (val int64, ok bool) {
	return i.value, true
}

func (i *integer) Marshal(w io.Writer) error {
	_, err := fmt.Fprintf(w, ":%d\r\n", i.value)
	return err
}

func (i *integer) Unmarshal(r ReaderPeeker) error {
	data, err := readUntilCRLF(r)
	if err != nil {
		return err
	}
	if len(data) < 1 || data[0] != prefixInteger {
		return fmt.Errorf("%w: simple string must start with '+'", ErrInvalidSyntax)
	}
	i.value, err = strconv.ParseInt(string(data[1:]), 10, 64)
	return err
}

type respError struct {
	baseValue[string]
}

func Error(value string) Value {
	return &respError{baseValue[string]{value: value}}
}

func (e *respError) Type() ValueType {
	return TypeError
}

func (e *respError) Error() (val string, ok bool) {
	return e.value, true
}

func (e *respError) Marshal(w io.Writer) error {
	if strings.Index(e.value, "\r\n") != -1 {
		return fmt.Errorf("%w: error contains CRLF", ErrMarshal)
	}
	_, err := fmt.Fprintf(w, "-%s\r\n", e.value)
	return err
}

func (e *respError) Unmarshal(r ReaderPeeker) error {
	data, err := readUntilCRLF(r)
	if err != nil {
		return err
	}
	if len(data) < 1 || data[0] != prefixError {
		return fmt.Errorf("%w: error string must start with '-'", ErrInvalidSyntax)
	}
	e.value = string(data[1:])
	return nil
}

type bulkString struct {
	baseValue[[]byte]
	isNull bool
}

func BulkString(value []byte) Value {
	return &bulkString{
		baseValue: baseValue[[]byte]{value: value},
		isNull:    false,
	}
}

func NullBulkString() Value {
	return &bulkString{
		baseValue: baseValue[[]byte]{value: nil},
		isNull:    true,
	}
}

func (s *bulkString) Type() ValueType {
	return TypeBulkString
}

func (s *bulkString) IsNull() bool {
	return s.isNull
}

func (s *bulkString) String() (val string, ok bool) {
	return string(s.value), true
}

func (s *bulkString) Bytes() (val []byte, ok bool) {
	return s.value, true
}

func (s *bulkString) Marshal(w io.Writer) error {
	_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(s.value), s.value)

	return err
}

func (s *bulkString) Unmarshal(r ReaderPeeker) error {
	rawHeader, err := readUntilCRLF(r)
	if err != nil {
		return err
	}
	if rawHeader[0] != prefixBulkString {
		return fmt.Errorf("%w: bulk string must start with '$'", ErrInvalidSyntax)
	}

	size, err := strconv.ParseInt(string(rawHeader[1:]), 10, 64)
	if err != nil {
		return err
	}

	if size == -1 {
		s.value = nil
		s.isNull = true
		return nil
	}
	s.isNull = false

	if size == 0 {
		s.value = nil
		return nil
	}

	rawData := make([]byte, size+2)
	if n, err := r.Read(rawData); err != nil || n != len(rawData) {
		return fmt.Errorf("%w: can't read enoguh data", ErrInvalidSyntax)
	}

	if string(rawData[len(rawData)-2:]) != "\r\n" {
		return fmt.Errorf("%w: not null bulk string must has crlf ending", ErrInvalidSyntax)
	}
	s.value = rawData[:len(rawData)-2]
	return nil
}

func unmarshalAny(r ReaderPeeker) (Value, error) {
	prefix, err := r.Peek(1)
	if err != nil {
		return nil, err
	}
	var value Value
	switch prefix[0] {
	case prefixSimpleString:
		value = &simpleString{}
	case prefixArray:
		value = &array{}
	case prefixInteger:
		value = &integer{}
	case prefixError:
		value = &respError{}
	case prefixBulkString:
		value = &bulkString{}
	default:
		return nil, fmt.Errorf("%w: unknown prefix %s", ErrMarshal, prefix)
	}

	err = value.Unmarshal(r)
	return value, err
}

func Unmarshal(r ReaderPeeker) (Value, error) {
	return unmarshalAny(r)
}

func Marshal(w io.Writer, v Value) error {
	writer := bufio.NewWriter(w)
	if err := v.Marshal(writer); err != nil {
		return err
	}
	return writer.Flush()
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
