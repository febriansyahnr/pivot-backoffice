package util

func SliceExtractOrDefault[T comparable](s []T, index int, d T) T {
	if len(s) > index {
		return s[index]
	}
	return d
}
