package goapi

import "reflect"

func toPtr[T any](v T) *T {
	return &v
}

func isFixedType(fType reflect.Type) bool {
	if fType == typeFile || fType == typeFiles || fType == typeCookie {
		return true
	}
	return false
}

func inArray[T comparable](val T, list []T) bool {
	for _, v := range list {
		if val == v {
			return true
		}
	}
	return false
}
