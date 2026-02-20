package tasklib

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func parseQuery(query map[string][]string, header map[string][]string, value any) error {
	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return errors.New("Must recieve a pointer to a struct")
	}

	stc := v.Elem()
	typ := stc.Type()
	i := 0
	for i < stc.NumField() {
		val := stc.Field(i)
		field := typ.Field(i)

		if !field.IsExported() || !val.CanSet() {
			i++
			continue
		}

		qVal := field.Tag.Get("query")
		if qVal != "" {
			q, ok := query[qVal]
			if !ok {
				i++
				continue
			}

			err := mapStringValue(val, q)
			if err != nil {
				return err
			}
		}

		hVal := field.Tag.Get("header")
		if hVal != "" {
			h, ok := header[qVal]
			if !ok {
				i++
				continue
			}

			err := mapStringValue(val, h)
			if err != nil {
				return err
			}
		}

		i++
	}

	return nil
}

func mapStringValue(val reflect.Value, value []string) error {
	lastValue := value[len(value)-1]
	switch val.Kind() {
	case reflect.String:
		val.SetString(strings.Join(value, " "))
	case reflect.Slice:
		if val.Type().Elem().Kind() != reflect.String {
			return errors.New("only []string supported")
		}
		val.Set(reflect.ValueOf(value))
	case reflect.Bool:
		b, err := strconv.ParseBool(lastValue)
		if err != nil {
			return err
		}
		val.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(lastValue, 10, val.Type().Bits())
		if err != nil {
			return err
		}
		val.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(lastValue, 10, val.Type().Bits())
		if err != nil {
			return err
		}
		val.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(lastValue, val.Type().Bits())
		if err != nil {
			return err
		}
		val.SetFloat(f)
	}
	return nil
}
