package gnuflag

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

// flagName takes a prefix, and a variable name and produces a "prefix-flag-name-with-dashes".
// It is intended to also detect acronyms all in upper case, as is Go style.
func flagName(prefix, name string) string {
	var words []string
	var word []rune
	var maybeAcronym bool

	if prefix != "" {
		words = append(words, prefix)
	}

	for _, r := range name {
		if unicode.IsUpper(r) {
			if !maybeAcronym && len(word) > 1 {
				words = append(words, string(word))
				word = word[:0]
			}

			maybeAcronym = true
			word = append(word, unicode.ToLower(r))
			continue
		}

		if maybeAcronym && len(word) > 1 {
			l := len(word) - 1

			words = append(words, string(word[:l]))

			word[0] = word[l]
			word = word[:1]
		}

		maybeAcronym = false
		word = append(word, r)
	}

	if len(word) > 0 {
		words = append(words, string(word))
	}

	return strings.Join(words, "-")
}

// arrayWrap is used to make array-based flags that will run the setterFunc on every element of a split on ","
func arrayWrap(fn setterFunc) setterFunc {
	return func(s string) error {
		for _, v := range strings.Split(s, ",") {
			if err := fn(v); err != nil {
				return err
			}
		}

		return nil
	}
}

// structVar is the work-horse, and does the actual reflection and recursive work.
func (fs *FlagSet) structVar(prefix string, v reflect.Value) error {
	typ := v.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		f := typ.Field(i)
		name := flagName(prefix, f.Name)

		usage := f.Tag.Get("desc")
		if usage == "" {
			usage = fmt.Sprintf("%s `%s`", f.Name, f.Type)
		}

		var short rune
		var defval string

		if tag := f.Tag.Get("flag"); tag != "" {
			fields := strings.Split(tag, ",")

			if len(fields) < 1 {
				continue
			}

			if fields[0] != "" {
				name = fields[0]
			}

			if name == "-" {
				continue
			}

			for _, field := range fields[1:] {
				switch {
				case strings.HasPrefix(field, "short="):
					// This is kind of a “cheat”, ranging over a string uses UTF-8 runes.
					// so we grab the first rune, and then break. No need to use utf8 package.
					for _, r := range field[len("short="):] {
						short = r
						break
					}

				case strings.HasPrefix(field, "default="):
					defval = field[len("default="):]
					if i+1 < len(fields) {
						// Commas aren't escaped, and def is always last.
						defval += "," + strings.Join(fields[i+1:], ",")
						break
					}

				case strings.HasPrefix(field, "def="):
					defval = field[4:]
					if i+1 < len(fields) {
						// Commas aren't escaped, and def is always last.
						defval += "," + strings.Join(fields[i+1:], ",")
						break
					}
				}
			}
		}

		var opts []Option
		if short != 0 {
			opts = append(opts, WithShort(short))
		}
		if defval != "" {
			opts = append(opts, WithDefault(defval))
		}

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				// if the pointer is nil, then allocate the appropriate type
				// and assign it into the pointer.
				p := reflect.New(f.Type.Elem())
				field.Set(p)
			}

			if _, ok := field.Interface().(Value); !ok {
				// now then, if we don’t implement Value, lets work with the element itself.
				field = field.Elem()
			}
		}

		// We set val such that we can generically just use fs.Var to setup the flag,
		// any other fs.TypeVar will overwrite the value that is stored in that field,
		// which means we wouldn’t get that value as the default.
		// But we want the value in the field as default, even if no `flag:",default=val"` is given.
		var val Value
		ptr := unsafe.Pointer(field.UnsafeAddr())

		if field.Kind() != reflect.Ptr {
			if _, ok := field.Interface().(Value); !ok {
				f := field.Addr()

				if _, ok := f.Interface().(Value); ok {
					field = f
				}
			}
		}

		switch v := field.Interface().(type) {
		case EnumValue:
			set := &enumValue{
				val: (*int)(ptr),
			}
			val = set

			if tag := f.Tag.Get("values"); tag != "" {
				set.setValid(strings.Split(tag, ","))
			}

		case Value:
			// this is obviously the simplest option… the work is already done.
			val = v

		case bool:
			val = (*boolValue)(ptr)

		case uint:
			val = (*uintValue)(ptr)
		case []uint:
			slice := (*[]uint)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				u, err := strconv.ParseUint(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, uint(u))
				return nil
			}))

		case uint64:
			val = (*uint64Value)(ptr)
		case []uint64:
			slice := (*[]uint64)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				u, err := strconv.ParseUint(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, u)
				return nil
			}))

		case uint8, uint16, uint32:
			// here we support a few additional types with generic-ish reflection
			val = newFunc(fmt.Sprint(field), func(s string) error {
				u, err := strconv.ParseUint(s, 0, 64)
				if err != nil {
					return err
				}

				field.SetUint(u)
				return nil
			})

		case int:
			val = (*intValue)(ptr)
		case []int:
			slice := (*[]int)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, int(i))
				return nil
			}))

		case int64:
			val = (*int64Value)(ptr)
		case []int64:
			slice := (*[]int64)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, i)
				return nil
			}))

		case int8, int16, int32:
			// here we support a few additional types with generic-ish reflection
			val = newFunc(fmt.Sprint(field), func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}

				field.SetInt(i)
				return nil
			})

		case float64:
			val = (*float64Value)(ptr)
		case []float64:
			slice := (*[]float64)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, f)
				return nil
			}))

		case float32:
			// here we support float32 with generic-ish reflection
			val = newFunc(fmt.Sprint(field), func(s string) error {
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return err
				}

				field.SetFloat(f)
				return nil
			})

		case string:
			val = (*stringValue)(ptr)
		case []string:
			slice := (*[]string)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				*slice = append(*slice, s)
				return nil
			}))

		case []byte:
			// just like string, but stored as []byte
			val = newFunc(fmt.Sprint(field), func(s string) error {
				field.SetBytes([]byte(s))
				return nil
			})

		case time.Duration:
			val = (*durationValue)(ptr)
		case []time.Duration:
			slice := (*[]time.Duration)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return err
				}

				*slice = append(*slice, d)
				return nil
			}))

		case url.URL:
			set := (*url.URL)(ptr)

			val = newFunc(fmt.Sprint(field), func(s string) error {
				uri, err := url.Parse(s)
				if err != nil {
					return err
				}

				*set = *uri
				return nil
			})
		case []*url.URL:
			slice := (*[]*url.URL)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			val = newFunc(def, arrayWrap(func(s string) error {
				uri, err := url.Parse(s)
				if err != nil {
					return err
				}

				*slice = append(*slice, uri)
				return nil
			}))

		default:
			if field.Kind() == reflect.Struct {
				fs.structVar(name, field)
				continue
			}

			panic(fmt.Sprintf("gnuflag: unsupported type %s for %s", f.Type, f.Name))
		}

		fs.Var(val, name, usage, opts...)
	}

	return nil
}

// Struct uses reflection to take a structure and turn it into a series of flags.
// It recognizes the struct tags of `flag:"flag-name,short=F,default=defval"` and `desc:"usage"`.
// The "desc" tag is intended to be much more generic than just for use in this library.
// To ignore a struct value use the tag `flag:"-"`, and `flag:","` will use the variable’s name.
func (fs *FlagSet) Struct(prefix string, value interface{}) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		panic(fmt.Sprintf("gnuflag.Flags on non-pointer: %v", v.Kind()))
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("gnuflag.Flags on non-struct: %v", v.Kind()))
	}

	return fs.structVar(prefix, v)
}

// Struct uses default CommandLine flagset.
func Struct(prefix string, value interface{}) error {
	return CommandLine.Struct(prefix, value)
}
