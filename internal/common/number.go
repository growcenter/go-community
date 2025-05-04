package common

import (
	"fmt"
	"reflect"
)

func LengthOf(v interface{}) (int, error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String, reflect.Chan:
		return rv.Len(), nil
	default:
		return 0, fmt.Errorf("type %T does not support len", v)
	}
}
