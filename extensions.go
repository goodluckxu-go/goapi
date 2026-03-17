package goapi

import (
	"reflect"
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

// GetConvert the extended content
// key prefix has 'x-'
func (l *Extensions) GetConvert(key string) Convert {
	return Convert(l.GetString(key))
}
