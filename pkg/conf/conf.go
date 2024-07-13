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
	ErrRequired         = errors.New("field is required")
)

type Setter interface {
	SetValue(value []string) error
}

func BindFile(cfg interface{}, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	return bindConfig(cfg, file)
}

func Bind(cfg interface{}, r io.Reader) error {
	return bindConfig(cfg, r)
}

func bindConfig(cfg interface{}, r io.Reader) error {
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		return fmt.Errorf("%w: config structure must be a pointer", ErrTypeNotSupported)
	}

	meta := configMeta{fields: make(map[string]field)}

	str := reflect.ValueOf(cfg).Elem()

	if err := collectStructMeta(&meta, str, ""); err != nil {
		return err
	}

	data := make(map[string][]string)
	if err := parse(data, r); err != nil {
		return err
	}

	for k, f := range meta.fields {
		val, ok := data[k]
		if !ok {
			if f.defaultValue != nil {
				val = []string{*f.defaultValue}
			} else if f.required {
				return fmt.Errorf("%w: key %s is required", ErrRequired, k)
			}
		}
		if err := bindValue(f.value, val); err != nil {
			return err
		}
	}
	return nil
}

func collectStructMeta(meta *configMeta, v reflect.Value, prefix string) error {
	t := v.Type()
	if t.Kind() != reflect.Struct {
		panic("value must be a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)
		if err := collectFieldMeta(meta, ft, fv, prefix); err != nil {
			return err
		}
	}

	return nil
}

func collectFieldMeta(meta *configMeta, f reflect.StructField, v reflect.Value, prefix string) error {
	fieldMeta := field{}
	_, fieldMeta.required = f.Tag.Lookup("redis-required")
	name, ok := f.Tag.Lookup("redis")
	if !ok {
		name = f.Name
	}
	fieldMeta.name = prefix + name
	fieldMeta.value = v
	defaultValue, ok := f.Tag.Lookup("redis-default")
	prefix += f.Tag.Get("redis-prefix")
	if ok {
		fieldMeta.defaultValue = &defaultValue
	}

	switch f.Type.Kind() {
	case reflect.Struct:
		return collectStructMeta(meta, v, prefix)
	case reflect.Slice:
		panic("slices are not supported")
	case reflect.Ptr:
		panic("pointers are not supported")
	default:
		meta.fields[fieldMeta.name] = fieldMeta
		return nil
	}
}

type configMeta struct {
	fields map[string]field
}

type field struct {
	name         string
	required     bool
	value        reflect.Value
	defaultValue *string
}

func bindValue(v reflect.Value, raw []string) error {
	if s, ok := v.Interface().(Setter); ok {
		return s.SetValue(raw)
	}
	if len(raw) == 0 {
		return fmt.Errorf("%w: can't bind empty value", ErrSyntax)
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(raw[0])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(raw[0], 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(raw[0], 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw[0], 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw[0])
		if err != nil {
			return err
		}
		v.SetBool(b)
	default:
		return fmt.Errorf("%w: %s", ErrTypeNotSupported, v.Type().String())
	}
	return nil
}
