package goapi

import (
	"strconv"
)

type Extensions struct {
	extensions map[string]any
	prefix     string
}

func (l *Extensions) new() (t *Extensions) {
	if len(l.extensions) == 0 {
		return l
	}
	t = new(Extensions)
	t.extensions = l.extensions
	t.prefix = l.prefix
	return
}

// Root Obtain the root level
func (l *Extensions) Root() (t *Extensions) {
	if len(l.extensions) == 0 {
		return l
	}
	t = l.new()
	t.prefix = ""
	return
}

// Struct kind of struct
func (l *Extensions) Struct(field string) (t *Extensions) {
	if len(l.extensions) == 0 {
		return l
	}
	t = l.new()
	name := "Struct_" + field
	if t.prefix == "" {
		t.prefix = name
	} else {
		t.prefix += "." + name
	}
	return
}

// Map kind of map
func (l *Extensions) Map() (t *Extensions) {
	if len(l.extensions) == 0 {
		return l
	}
	t = l.new()
	name := "Map"
	if t.prefix == "" {
		t.prefix = name
	} else {
		t.prefix += "." + name
	}
	return
}

// Slice kind of array/slice
func (l *Extensions) Slice() (t *Extensions) {
	if len(l.extensions) == 0 {
		return l
	}
	t = l.new()
	name := "Slice"
	if t.prefix == "" {
		t.prefix = name
	} else {
		t.prefix += "." + name
	}
	return
}

// Get the extended content
// key prefix has 'x-'
func (l *Extensions) Get(key string) (val any, ok bool) {
	if len(l.extensions) == 0 {
		return
	}
	val, ok = l.extensions[l.prefix+"."+key]
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
