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

// flagName takes a variable name and produces from "FlagNameWithDashes" a "flag-name-with-dashes".
// It also attempts to detect acronyms all in upper case, as is Go style.
// It is _best effort_ and may entirely mangle your flagname,
// e.g. "HTTPURL" will be interpreted as a single acronym.
// It exists solely to give a better default than strings.ToLower(),
// it should almost certainly be overridden with an intentional name.
func flagName(name string) string {
	var words []string
	var word []rune
	var maybeAcronym bool

	for _, r := range name {
		switch {
		case unicode.IsUpper(r):
			if !maybeAcronym && len(word) > 0 {
				words = append(words, string(word))
				word = word[:0] // reuse the previous allocated slice.
			}

			maybeAcronym = true
			word = append(word, unicode.ToLower(r))

		case maybeAcronym && len(word) > 1: // an acronym can only be from two uppercase letters together.
			l := len(word) - 1

			words = append(words, string(word[:l]))

			word[0] = word[l]
			word = word[:1]
			fallthrough

		default:
			maybeAcronym = false
			word = append(word, r)
		}
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
	if strings.Contains(prefix, "=") || strings.HasPrefix(prefix, "-") {
		return fmt.Errorf("invalid prefix: %q", prefix)
	}

	structType := v.Type()

	for i := 0; i < structType.NumField(); i++ {
		val := v.Field(i)
		if !val.CanSet() {
			continue
		}

		field := structType.Field(i)
		name := flagName(field.Name)

		usage := field.Tag.Get("desc")
		if usage == "" {
			usage = fmt.Sprintf("%s `%s`", field.Name, field.Type)
		}

		var short rune
		defval := field.Tag.Get("default")

		if tag := field.Tag.Get("flag"); tag != "" {
			directives := strings.Split(tag, ",")

			if len(directives) >= 1 {
				// sanity check: by documentation this should always be true.

				if directives[0] != "" {
					name = directives[0]
				}
			}

			if name == "-" {
				continue
			}

		directivesLoop:
			for j := 1; j < len(directives); j++ {
				directive := directives[j]

				switch {
				case strings.HasPrefix(directive, "short="):
					// This is kind of a “cheat”, ranging over a string uses UTF-8 runes.
					// so we grab the first rune, and then break. No need to use utf8 package.
					for _, r := range directive[len("short="):] {
						short = r
						break
					}

				case strings.HasPrefix(directive, "default="):
					defval = strings.TrimPrefix(directive, "default=")
					if j+1 < len(directives) {
						// Commas aren't escaped, and default is defined to be last.
						defval += "," + strings.Join(directives[j+1:], ",")
						break directivesLoop
					}

				case strings.HasPrefix(directive, "def="):
					defval = strings.TrimPrefix(directive, "def=")
					if j+1 < len(directives) {
						// Commas aren't escaped, and def is defined to be last.
						defval += "," + strings.Join(directives[j+1:], ",")
						break directivesLoop
					}
				}
			}
		}

		if prefix != "" {
			name = prefix + "-" + name
		}

		if strings.Contains(name, "=") || strings.HasPrefix(name, "-") {
			return fmt.Errorf("invalid flag name for field %s: %s", field.Name, name)
		}

		var opts []Option
		if short != 0 {
			opts = append(opts, WithShort(short))
		}
		if defval != "" {
			opts = append(opts, WithDefault(defval))
		}

		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				// if the pointer is nil, then allocate the appropriate type
				// and assign it into the pointer.
				p := reflect.New(val.Type().Elem())
				val.Set(p)
			}

			if _, ok := val.Interface().(Value); !ok {
				// if val does not implement flag.Value, let's work with the element itself.
				val = val.Elem()
			}
		}

		// We set value such that we can generically just use fs.Var to setup the flag,
		// any other FlagSet.TypeVar will overwrite the value that is stored in that field,
		// which means we wouldn’t get that value as the default.
		// But we want the value in the field as default, even if no `flag:",default=val"` is given.
		var value Value

		// We reference the value in the struct directly in order to properly be able to set its value,
		// and this is the only way to get that.
		ptr := unsafe.Pointer(val.UnsafeAddr())

		if val.Kind() != reflect.Ptr {
			// if the value is not a pointer:

			if _, ok := val.Interface().(Value); !ok {
				// and the value itself does not implement flag.Value:
				pval := val.Addr()

				if _, ok := pval.Interface().(Value); ok {
					// but a pointer to the value does implement flag.Value:
					// then use that value instead.
					val = pval
				}
			}
		}

		switch v := val.Interface().(type) {
		case EnumValue:
			enum := &enumValue{
				val: (*int)(ptr),
			}
			value = enum

			if tag := field.Tag.Get("values"); tag != "" {
				enum.setValid(strings.Split(tag, ","))
			}

		case Value:
			// this is obviously the simplest option… the work is already done.
			value = v

		case bool:
			value = (*boolValue)(ptr)

		case uint:
			value = (*uintValue)(ptr)
		case []uint:
			slice := (*[]uint)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			value = newFunc(def, arrayWrap(func(s string) error {
				u, err := strconv.ParseUint(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, uint(u))
				return nil
			}))

		case uint64:
			value = (*uint64Value)(ptr)
		case []uint64:
			slice := (*[]uint64)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			value = newFunc(def, arrayWrap(func(s string) error {
				u, err := strconv.ParseUint(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, u)
				return nil
			}))

		case uint8, uint16, uint32:
			// here we support a few additional types with generic-ish reflection
			value = newFunc(fmt.Sprint(field), func(s string) error {
				u, err := strconv.ParseUint(s, 0, 64)
				if err != nil {
					return err
				}

				val.SetUint(u)
				return nil
			})

		case int:
			value = (*intValue)(ptr)
		case []int:
			slice := (*[]int)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			value = newFunc(def, arrayWrap(func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, int(i))
				return nil
			}))

		case int64:
			value = (*int64Value)(ptr)
		case []int64:
			slice := (*[]int64)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			value = newFunc(def, arrayWrap(func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, i)
				return nil
			}))

		case int8, int16, int32:
			// here we support a few additional types with generic-ish reflection
			value = newFunc(fmt.Sprint(field), func(s string) error {
				i, err := strconv.ParseInt(s, 0, 64)
				if err != nil {
					return err
				}

				val.SetInt(i)
				return nil
			})

		case float64:
			value = (*float64Value)(ptr)
		case []float64:
			slice := (*[]float64)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			value = newFunc(def, arrayWrap(func(s string) error {
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return err
				}

				*slice = append(*slice, f)
				return nil
			}))

		case float32:
			// here we support float32 with generic-ish reflection
			value = newFunc(fmt.Sprint(field), func(s string) error {
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return err
				}

				val.SetFloat(f)
				return nil
			})

		case string:
			value = (*stringValue)(ptr)
		case []string:
			slice := (*[]string)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			value = newFunc(def, arrayWrap(func(s string) error {
				*slice = append(*slice, s)
				return nil
			}))

		case []byte:
			// just like string, but stored as []byte
			value = newFunc(fmt.Sprint(field), func(s string) error {
				val.SetBytes([]byte(s))
				return nil
			})

		case time.Duration:
			value = (*durationValue)(ptr)
		case []time.Duration:
			slice := (*[]time.Duration)(ptr)

			var def string
			if len(*slice) > 0 {
				def = fmt.Sprint(*slice)
			}

			value = newFunc(def, arrayWrap(func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return err
				}

				*slice = append(*slice, d)
				return nil
			}))

		case url.URL:
			set := (*url.URL)(ptr)

			value = newFunc(fmt.Sprint(field), func(s string) error {
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

			value = newFunc(def, arrayWrap(func(s string) error {
				uri, err := url.Parse(s)
				if err != nil {
					return err
				}

				*slice = append(*slice, uri)
				return nil
			}))

		default:
			if val.Kind() != reflect.Struct {
				return fmt.Errorf("gnuflag: unsupported type %s for %s", field.Type, field.Name)
			}

			if err := fs.structVar(name, val); err != nil {
				return err
			}

			continue
		}

		if err := fs.Var(value, name, usage, opts...); err != nil {
			return err
		}
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
		return fmt.Errorf("gnuflag.FlagSet.Struct on non-pointer: %v", v.Kind())
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("gnuflag.FlagSet.Struct on non-struct: %v", v.Kind())
	}

	return fs.structVar(prefix, v)
}

// Struct uses default CommandLine flagset.
func Struct(prefix string, value interface{}) error {
	return CommandLine.Struct(prefix, value)
}
