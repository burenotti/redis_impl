package conf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

var (
	ErrNotFound         = errors.New("key not found")
	ErrTypeNotSupported = errors.New("type not supported")
)

type Config struct {
	data map[string][]string
}

func read(r io.Reader) (*Config, error) {
	c := Config{
		data: make(map[string][]string),
	}
	err := parse(c.data, r)
	return &c, err
}

func Read(r io.Reader) (*Config, error) {
	return read(r)
}

func ReadFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return read(file)
}

func (c *Config) Get(key string) Field {
	return Field{
		config: c,
		key:    key,
		idx:    0,
	}
}

type Field struct {
	config *Config
	key    string
	idx    int
}

func (f Field) At(idx int) Field {
	return Field{
		config: f.config,
		key:    f.key,
		idx:    idx,
	}
}

func (f Field) String(or ...string) (string, error) {
	return value(&f, or...)
}

func (f Field) MustString(or ...string) string {
	return must(value(&f, or...))
}

func (f Field) Int(or ...int) (int, error) {
	return value(&f, or...)
}

func (f Field) MustInt(or ...int) int {
	return must(value(&f, or...))
}

func (f Field) Float(or ...float64) (float64, error) {
	return value(&f, or...)
}

func (f Field) MustFloat(or ...float64) float64 {
	return must(value(&f, or...))
}

func (f Field) Len() int {
	val, ok := f.config.data[f.key]
	if !ok {
		return 0
	}
	return len(val)
}

func value[T any](f *Field, or ...T) (T, error) {
	if len(or) > 1 {
		panic("only one argument allowed")
	}

	val, ok := f.config.data[f.key]
	var null T
	if !ok {
		if len(or) == 0 {
			return null, fmt.Errorf("%w: %s", ErrNotFound, f.key)
		}
		return or[0], nil
	}

	if f.idx >= len(val) {
		if len(or) == 0 {
			return null, fmt.Errorf("%w: %s[%d]", ErrNotFound, f.key, f.idx)
		}
		return or[0], nil
	}

	return cast[T](val[f.idx])
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func cast[T any](value string) (res T, err error) {
	var v interface{}
	switch reflect.TypeOf(res).Kind() {
	case reflect.Int:
		v, err = strconv.Atoi(value)
	case reflect.String:
		v, err = value, nil
	case reflect.Float64:
		v, err = strconv.ParseFloat(value, 64)
	default:
		v, err = nil, fmt.Errorf("%w: %s", ErrTypeNotSupported, reflect.TypeOf(res))
	}

	return v.(T), err
}
