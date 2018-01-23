package util

import (
	"reflect"
)

func InterfaceSlice(value interface{}) []interface{} {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		var a []interface{}

		for i := 0; i < v.Len(); i++ {
			a = append(a, v.Index(i).Interface())
		}

		return a
	}

	panic("util.InterfaceSlice called with value other than array or slice")
}
