package goapi

import (
	"strconv"
)

type Extensions map[string][]string

// Get Obtain extended information, only tags at the same level as 'goapi.Router' are supported
//
//	struct {
//		router goapi.Router `paths:"/" methods:"get" x-test1:"1"` // x-test1 are supported
//		Body struct {
//			Page int `query:"page" x-error:"error"` // x-error no supported, not at the same level as 'goapi.Router'
//		} `x-test2:"2"` // x-test2 are supported
//		Limit int `query:"limit" x-test3:"3"` // x-test3 are supported
//	}
//
// key prefix has 'x-'
func (l Extensions) Get(key string) (val string) {
	list := l[key]
	if len(list) > 0 {
		val = list[0]
	}
	return
}

// GetInt the extended content
// key prefix has 'x-'
func (l Extensions) GetInt(key string) int64 {
	valStr := l.Get(key)
	if valStr == "" {
		return 0
	}
	val, _ := strconv.ParseInt(valStr, 10, 64)
	return val
}

// GetUint the extended content
// key prefix has 'x-'
func (l Extensions) GetUint(key string) uint64 {
	valStr := l.Get(key)
	if valStr == "" {
		return 0
	}
	val, _ := strconv.ParseUint(valStr, 10, 64)
	return val
}

// GetFloat the extended content
// key prefix has 'x-'
func (l Extensions) GetFloat(key string) float64 {
	valStr := l.Get(key)
	if valStr == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(valStr, 64)
	return val
}

// GetBool the extended content
// key prefix has 'x-'
func (l Extensions) GetBool(key string) bool {
	valStr := l.Get(key)
	if valStr == "" {
		return false
	}
	val, _ := strconv.ParseBool(valStr)
	return val
}
