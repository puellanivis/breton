package m3u8

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var quotedString = regexp.MustCompile(`^"(?:[^"]+|\\")*"`)

func unmarshalAttributeField(field reflect.Value, f reflect.StructField, val []byte) error {
	switch field.Interface().(type) {
	case Resolution:
		return field.Addr().Interface().(*Resolution).TextUnmarshal(val)

	case time.Time:
		t, err := time.Parse("2006-01-02T15:04:05.999Z07:00", string(val))
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(t))

	case time.Duration:
		s, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return err
		}

		d := time.Duration(s * float64(time.Second))
		field.Set(reflect.ValueOf(d))

	case bool:
		var b bool
		var err error

		values := f.Tag.Get("enum")
		switch values {
		case "":
			b, err = strconv.ParseBool(string(val))

		default:
			e := getEnum(values)
			i, err := e.Index(string(val))
			if err != nil {
				return err
			}

			if i > 0 {
				b = true
			}
		}

		if err != nil {
			return err
		}
		field.SetBool(b)

	case string:
		s := string(val)

		values := f.Tag.Get("enum")
		switch values {
		case "":
			s, err := strconv.Unquote(s)
			if err != nil {
				return err
			}

			field.SetString(s)

		default:
			e := getEnum(values)
			if _, err := e.Test(s); err != nil {
				return err
			}

			field.SetString(s)
		}

	case int, int8, int16, int32, int64:
		i, err := strconv.ParseInt(string(val), 10, 0)
		if err != nil {
			return err
		}

		field.SetInt(i)

	case uint, uint8, uint16, uint32, uint64:
		u, err := strconv.ParseUint(string(val), 10, 0)
		if err != nil {
			return err
		}

		field.SetUint(u)

	case float32:
		f, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case float64:
		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case []string:
		s, err := strconv.Unquote(string(val))
		if err != nil {
			return err
		}

		values := strings.Split(s, f.Tag.Get("delim"))

		field.Set(reflect.AppendSlice(field, reflect.ValueOf(values)))

	case []int:
		s, err := strconv.Unquote(string(val))
		if err != nil {
			return err
		}

		values := strings.Split(s, f.Tag.Get("delim"))

		var ints []int
		for _, value := range values {
			i, err := strconv.Atoi(value)
			if err != nil {
				return err
			}

			ints = append(ints, i)
		}
		field.Set(reflect.AppendSlice(field, reflect.ValueOf(ints)))

	case []byte:
		s := string(val)

		if !strings.HasPrefix(s, "0x") {
			return fmt.Errorf("hexidecimal-sequence does not start with 0x")
		}

		b, err := hex.DecodeString(s[2:])
		if err != nil {
			return err
		}

		field.SetBytes(b)

	default:
		return fmt.Errorf("unknown attribute-list field of type %T", field.Interface())
	}

	return nil
}

func unmarshalAttributeList(val interface{}, value []byte) error {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("m3u8.unmarshalAttributeList on non-pointer: %v", v.Kind())
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("m3u8.unmarshalAttributeList on non-struct pointer: %v", v.Kind())
	}

	typ := v.Type()

	var values [][]byte

	for len(value) > 0 {
		i := bytes.IndexAny(value, "=,")
		if i < 0 {
			values = append(values, value[0:len(value):len(value)])
			break
		}

		if value[i] == '=' {
			i++

			var inQuotes bool
			for ; i < len(value); i++ {
				if value[i] == '"' {
					inQuotes = !inQuotes
					continue
				}

				if !inQuotes && value[i] == ',' {
					break
				}

				if inQuotes && value[i] == '\\' {
					i++
				}
			}
		}

		values = append(values, value[0:i:i])
		if i == len(value) {
			value = nil
			continue
		}
		value = value[i+1:]
	}

	for _, value := range values {
		var wasSet bool

		for i := 0; i < typ.NumField(); i++ {
			field := v.Field(i)
			if !field.CanSet() {
				continue
			}

			f := typ.Field(i)
			name := strings.ToUpper(f.Name)

			if tag := f.Tag.Get("m3u8"); tag != "" {
				fields := strings.Split(tag, ",")

				if fields[0] != "" {
					name = fields[0]
				}

				if name == "-" {
					continue
				}
			}

			if !bytes.HasPrefix(value, []byte(name)) {
				continue
			}

			value = value[len(name):]
			if value[0] == '=' {
				value = value[1:]
			}

			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					p := reflect.New(f.Type.Elem())
					field.Set(p)
				}

				field = field.Elem()
			}

			switch field.Interface().(type) {
			case map[string]interface{}:
				if field.IsNil() {
					m := reflect.MakeMap(f.Type)
					field.Set(m)
				}

				i := bytes.IndexByte(value, '=')
				key := reflect.ValueOf(string(value[:i]))

				var val interface{}

				v := string(value[i+1:])

				if strings.HasPrefix(v, "0x") {
					b, err := hex.DecodeString(v[2:])
					if err != nil {
						return err
					}
					val = b

				} else if s, err := strconv.Unquote(v); err == nil {
					val = s

				} else if f, err := strconv.ParseFloat(v, 64); err == nil {
					val = f

				} else {
					return fmt.Errorf("%s: invalid client-attribute: %s", key, v)
				}

				field.SetMapIndex(key, reflect.ValueOf(val))
				wasSet = true
				continue
			}

			err := unmarshalAttributeField(field, f, value)
			if err != nil {
				return fmt.Errorf("%s: %v", name, err)
			}

			wasSet = true
		}

		if !wasSet {
			return fmt.Errorf("unknown attribute-list field: %q", value)
		}
	}

	return nil
}

func marshalAttributeField(field reflect.Value, f reflect.StructField) (s string, omit bool, err error) {
	switch v := field.Interface().(type) {
	case Resolution:
		omit = (v.Width + v.Height) == 0
		s = v.String()

	case time.Time:
		omit = v.IsZero()

		format := f.Tag.Get("format")
		switch format {
		case "":
			s = v.String()
		default:
			s = v.Format(format)
		}

	case time.Duration:
		f := v.Seconds()

		omit = v == 0
		s = strconv.FormatFloat(f, 'f', -1, 64)

	case fmt.Stringer:
		s := v.String()
		omit = s == ""
		s = strconv.Quote(s)

	case bool:
		omit = !v

		values := f.Tag.Get("enum")
		switch values {
		case "":
			s = strconv.FormatBool(v)

		default:
			e := getEnum(values)
			i := 0
			if v {
				i = 1
			}
			s, _ = e.Value(i)
		}

	case string:
		omit = v == ""

		values := f.Tag.Get("enum")
		switch values {
		case "":
			s = strconv.Quote(v)

		default:
			e := getEnum(values)
			s, err = e.Test(v)
		}

	case int, int8, int16, int32, int64:
		i := field.Int()

		omit = i == 0
		s = fmt.Sprintf("%d", i)

	case uint, uint8, uint16, uint32, uint64:
		u := field.Uint()

		omit = u == 0
		s = fmt.Sprintf("%d", u)

	case float32:
		omit = v == 0
		s = strconv.FormatFloat(float64(v), 'f', -1, 32)

	case float64:
		omit = v == 0
		s = strconv.FormatFloat(v, 'f', -1, 64)

	case []string:
		omit = len(v) == 0
		s = strconv.Quote(strings.Join(v, f.Tag.Get("delim")))

	case []int:
		omit = len(v) == 0

		var fields []string
		for _, field := range v {
			fields = append(fields, fmt.Sprint(field))
		}

		s = strconv.Quote(strings.Join(fields, f.Tag.Get("delim")))

	case []byte:
		omit = len(v) == 0
		s = hex.EncodeToString(v)

	default:
		return "", false, fmt.Errorf("unknown attribute-list field of type %T", v)
	}

	return s, omit, err
}

func marshalAttributeList(val interface{}) (string, error) {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return "", fmt.Errorf("m3u8.unmarshalAttributeList on non-pointer: %v", v.Kind())
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("m3u8.unmarshalAttributeList on non-struct pointer: %v", v.Kind())
	}

	typ := v.Type()

	var list []string

	for i := 0; i < typ.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		f := typ.Field(i)
		name := strings.ToUpper(f.Name)

		var optional bool

		if tag := f.Tag.Get("m3u8"); tag != "" {
			fields := strings.Split(tag, ",")

			if fields[0] != "" {
				name = fields[0]
			}

			if name == "-" {
				continue
			}

			for _, field := range fields[1:] {
				switch {
				case field == "optional":
					optional = true
				}
			}
		}

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				p := reflect.New(f.Type.Elem())
				field.Set(p)
			}

			if _, ok := field.Interface().(fmt.Stringer); !ok {
				field = field.Elem()
			}
		}

		switch v := field.Interface().(type) {
		case map[string]interface{}:
			if len(v) == 0 {
				continue
			}

			var keys []string
			for key := range v {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				field := v[key]
				var s string

				switch v := field.(type) {
				case string:
					s = strconv.Quote(v)
				case int, int8, int16, int32, int64:
					s = fmt.Sprintf("%d", v)
				case uint, uint8, uint16, uint32, uint64:
					s = fmt.Sprintf("%d", v)
				case float32:
					s = strconv.FormatFloat(float64(v), 'f', -1, 32)
				case float64:
					s = strconv.FormatFloat(float64(v), 'f', -1, 64)
				case []byte:
					s = fmt.Sprintf("0x%02X", v)
				default:
					return "", fmt.Errorf("unknown client-attribute field %s of type %T", key, v)
				}

				list = append(list, fmt.Sprintf("%s%s=%s", name, key, s))
			}
			continue
		}

		s, omit, err := marshalAttributeField(field, f)
		if err != nil {
			return "", err
		}

		if omit && optional {
			continue
		}

		list = append(list, fmt.Sprintf("%s=%s", name, s))
	}

	return strings.Join(list, ","), nil
}
