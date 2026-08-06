package encoder

import (
	"net/http"
	"reflect"
	"strings"
)

var canNilCheck = map[reflect.Kind]bool{
	reflect.Slice:     true,
	reflect.Interface: true,
	reflect.Map:       true,
	reflect.Ptr:       true,
}

type Inspector interface {
	Inspects(data interface{}) interface{}
}

func NewInspector(fields []string) Inspector {
	e := &encoder{
		disguisedKeys: map[string]bool{},
	}
	for _, field := range fields {
		e.disguisedKeys[strings.ToLower(field)] = true
	}
	return e
}

func (e *encoder) Inspects(data interface{}) interface{} {
	if len(e.disguisedKeys) == 0 || data == nil {
		return data
	}

	vo := reflect.Indirect(
		reflect.ValueOf(data),
	)

	switch vo.Type().Kind() {
	case reflect.Map:
		data = e.inspectMap(data)

	case reflect.Struct:
		data = e.inspectStruct(data)

	case reflect.Array, reflect.Slice:
		data = e.inspectSlice(data)
	}

	return data
}

func (e *encoder) inspectMap(data interface{}) interface{} {
	if data == nil {
		return data
	}

	src := reflect.Indirect(
		reflect.ValueOf(data),
	)
	maps := reflect.MakeMap(src.Type())

	for _, key := range src.MapKeys() {

		maps.SetMapIndex(key, src.MapIndex(key))
		if canNilCheck[maps.MapIndex(key).Type().Kind()] && maps.MapIndex(key).IsNil() {
			continue
		}

		val := reflect.Indirect(
			reflect.ValueOf(
				maps.MapIndex(key).Interface(),
			),
		)

		switch val.Type().Kind() {
		case reflect.Struct:
			maps.SetMapIndex(key, reflect.ValueOf(e.inspectStruct(val.Interface())))

		case reflect.String:
			maps.SetMapIndex(key, reflect.ValueOf(e.inspectValue(key.String(), val.String())))

		case reflect.Map:
			if tmp, ok := e.allowedMap(val.Interface()); ok {
				maps.SetMapIndex(key, reflect.ValueOf(e.inspectMap(tmp)))
			}

		case reflect.Slice, reflect.Array:
			if val.Len() == 0 {
				continue

			} else if val.Index(0).Type().Kind() == reflect.String {
				maps.SetMapIndex(key, reflect.ValueOf(e.inspectSliceStr(key.String(), val.Interface())))
			}
		}
	}
	return maps.Interface()
}

func (e *encoder) inspectStruct(data interface{}) interface{} {

	src := reflect.Indirect(
		reflect.ValueOf(data),
	)
	dst := reflect.New(src.Type()).Elem()
	dst.Set(src)

	to := dst.Type()

	for i := 0; i < to.NumField(); i++ {

		val, fl := dst.Field(i), to.Field(i)
		if canNilCheck[val.Type().Kind()] && val.IsNil() {
			continue
		}

		key := fl.Tag.Get("json")
		if secret := fl.Tag.Get("secret"); secret != "" {
			key = secret

		} else if key == "" {
			key = fl.Name
		}

		switch fl.Type.Kind() {
		case reflect.String:
			val.SetString(e.inspectValue(key, val.String()))

		case reflect.Map:
			val.Set(reflect.ValueOf(e.inspectMap(val.Interface())))

		case reflect.Slice, reflect.Array:
			if val.Len() == 0 {
				continue

			} else if val.Index(0).Type().Kind() == reflect.String {
				val.Set(reflect.ValueOf(e.inspectSliceStr(key, val.Interface())))
			}

		case reflect.Struct:
			val.Set(reflect.ValueOf(e.inspectStruct(val.Interface())))

		case reflect.Interface:
			val.Set(reflect.ValueOf(e.Inspects(val.Interface())))
		}
	}
	return dst.Interface()
}

func (e *encoder) inspectSlice(data interface{}) interface{} {
	if data == nil {
		return data
	}

	src := reflect.Indirect(
		reflect.ValueOf(data),
	)
	dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())

	for i := 0; i < src.Len(); i++ {

		dst.Index(i).Set(src.Index(i))
		if canNilCheck[dst.Index(i).Type().Kind()] && dst.Index(i).IsNil() {
			continue
		}

		val := reflect.Indirect(dst.Index(i))

		switch val.Type().Kind() {
		case reflect.Struct:
			val = reflect.ValueOf(e.inspectStruct(val.Interface()))

		case reflect.Map:
			val = reflect.ValueOf(e.inspectMap(val.Interface()))
		}

		if dst.Index(i).Type().Kind() == reflect.Ptr {
			dst.Index(i).Elem().Set(val)

		} else {
			dst.Index(i).Set(val)
		}
	}
	return dst.Interface()
}

func (e *encoder) inspectSliceStr(key string, values interface{}) interface{} {
	if !e.disguisedKeys[strings.ToLower(key)] {
		return values
	}

	src := reflect.Indirect(
		reflect.ValueOf(values),
	)

	var dst reflect.Value
	if src.Type().Kind() == reflect.Slice {
		dst = reflect.MakeSlice(src.Type(), src.Len(), src.Cap())

	} else {
		dst = reflect.New(src.Type()).Elem()
	}

	for i := 0; i < src.Len(); i++ {
		dst.Index(i).SetString(strings.Repeat("*", src.Index(i).Len()))
	}
	return dst.Interface()
}

func (e *encoder) inspectValue(key, val string) string {
	if e.disguisedKeys[strings.ToLower(key)] {
		return strings.Repeat("*", len(val))
	}
	return val
}

func (e *encoder) allowedMap(data interface{}) (interface{}, bool) {
	switch t := data.(type) {
	case map[string]interface{}:
		return t, true

	case map[string]string:
		return t, true

	case http.Header:
		return t, true

	default:
		return t, false
	}
}
