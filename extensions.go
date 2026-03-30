package goapi

import (
	"reflect"
	"strconv"
)

type Extensions struct {
	param   *paramField
	ins     []*inParam
	structs map[string]*structInfo
}

func (l *Extensions) init() {
	if l.param == nil {
		l.param = &paramField{
			kind: reflect.Struct,
			tag:  &paramTag{},
		}
		for _, in := range l.ins {
			l.param.fields = append(l.param.fields, in.field)
		}
	}
}

func (l *Extensions) new() (t *Extensions) {
	t = &Extensions{param: l.param, ins: l.ins, structs: l.structs}
	t.init()
	return
}

// Root Obtain the root level
func (l *Extensions) Root() (t *Extensions) {
	t = &Extensions{ins: l.ins, structs: l.structs}
	t.init()
	return
}

// Struct kind of struct
func (l *Extensions) Struct(field string) (t *Extensions) {
	t = l.new()
	if len(t.param.fields) == 0 {
		t.param.fields = t.structs[t.param.pkgName].fields
	}
	for _, pf := range t.param.fields {
		if pf.name == field {
			t.param = pf
			return t
		}
	}
	return nil
}

// Map kind of map
func (l *Extensions) Map() (t *Extensions) {
	t = l.new()
	t.param = t.param.fields[1]
	return
}

// Slice kind of array/slice
func (l *Extensions) Slice() (t *Extensions) {
	t = l.new()
	t.param = t.param.fields[0]
	return l
}

// Get the extended content
// key prefix has 'x-'
func (l *Extensions) Get(key string) (val any, ok bool) {
	l.init()
	val, ok = l.param.tag.extensions[key]
	return
}

// GetValue the extended content
// key prefix has 'x-'
func (l *Extensions) GetValue(key string) (val any) {
	val, _ = l.Get(key)
	return
}

// GetString the extended content
// key prefix has 'x-'
func (l *Extensions) GetString(key string) string {
	valStr, _ := l.GetValue(key).(string)
	return valStr
}

// GetInt the extended content
// key prefix has 'x-'
func (l *Extensions) GetInt(key string) int64 {
	valStr, ok := l.GetValue(key).(string)
	if !ok {
		return 0
	}
	val, _ := strconv.ParseInt(valStr, 10, 64)
	return val
}

// GetUint the extended content
// key prefix has 'x-'
func (l *Extensions) GetUint(key string) uint64 {
	valStr, ok := l.GetValue(key).(string)
	if !ok {
		return 0
	}
	val, _ := strconv.ParseUint(valStr, 10, 64)
	return val
}

// GetFloat the extended content
// key prefix has 'x-'
func (l *Extensions) GetFloat(key string) float64 {
	valStr, ok := l.GetValue(key).(string)
	if !ok {
		return 0
	}
	val, _ := strconv.ParseFloat(valStr, 64)
	return val
}

// GetBool the extended content
// key prefix has 'x-'
func (l *Extensions) GetBool(key string) bool {
	valStr, ok := l.GetValue(key).(string)
	if !ok {
		return false
	}
	val, _ := strconv.ParseBool(valStr)
	return val
}
