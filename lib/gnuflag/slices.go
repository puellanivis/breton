package gnuflag

import (
	"fmt"
	"reflect"
	"strings"
)

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

type sliceValue struct {
	value interface{} // this must be a pointer to the slice!
	fn    func(s string) error
}

func newSlice(value interface{}, fn func(s string) error) sliceValue {
	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("newSlice on non-pointer: %v", v.Kind()))
	}

	v = v.Elem()

	if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("newSlice on non-slice: %v", v.Kind()))
	}

	return sliceValue{
		value: value,
		fn:    fn,
	}
}

func (a sliceValue) Set(s string) error {
	for _, v := range strings.Split(s, ",") {
		if err := a.fn(v); err != nil {
			return err
		}
	}

	return nil
}

func (a sliceValue) Get() interface{} {
	v := reflect.ValueOf(a.value)

	return v.Elem().Interface()
}

func (a sliceValue) String() string {
	var elems []string

	val := reflect.ValueOf(a.value).Elem()

	for i := 0; i < val.Len(); i++ {
		v := val.Index(i)
		elems = append(elems, fmt.Sprint(v))
	}

	return strings.Join(elems, ",")
}
