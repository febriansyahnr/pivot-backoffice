package util

import "strings"

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func ContainsPrefix(slice []string, path string) bool {
	for _, item := range slice {
		if strings.Contains(path, item) {
			return true
		}
	}
	return false
}
