package util

func LookupMapAnyByKey[T comparable](data any, key string) (val T, found bool) {
	if data == nil {
		return
	}

	var src map[string]any

	switch tmp := data.(type) {
	case map[string]any:
		src = tmp

	case *map[string]any:
		src = *tmp

	default:
		return
	}

	if val, found = src[key].(T); !found {
		return
	}
	return val, true
}
