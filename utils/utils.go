package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func JoinString(args ...string) string {
	return strings.Join(args, "")
}

func ToString(i any) string {
	if i == nil {
		return ""
	}
	switch i.(type) {
	case string:
		return i.(string)
	case *string:
		return *i.(*string)
	case int, int8, int16, int32, int64:
		var i64 int64
		switch i.(type) {
		case int:
			i64 = int64(i.(int))
		case int8:
			i64 = int64(i.(int8))
		case int16:
			i64 = int64(i.(int16))
		case int32:
			i64 = int64(i.(int32))
		case int64:
			i64 = i.(int64)
		}
		return strconv.FormatInt(i64, 10)
	case *int, *int8, *int16, *int32, *int64:
		var i64 int64
		switch i.(type) {
		case *int:
			i64 = int64(*i.(*int))
		case *int8:
			i64 = int64(*i.(*int8))
		case *int16:
			i64 = int64(*i.(*int16))
		case *int32:
			i64 = int64(*i.(*int32))
		case *int64:
			i64 = *i.(*int64)
		}
		return strconv.FormatInt(i64, 10)
	case *uint, *uint8, *uint16, *uint32, *uint64:
		var ui64 uint64
		switch i.(type) {
		case *uint:
			ui64 = uint64(*i.(*uint))
		case *uint8:
			ui64 = uint64(*i.(*uint8))
		case *uint16:
			ui64 = uint64(*i.(*uint16))
		case *uint32:
			ui64 = uint64(*i.(*uint32))
		case *uint64:
			ui64 = *i.(*uint64)
		}
		return strconv.FormatUint(ui64, 10)
	case uint, uint8, uint16, uint32, uint64:
		var ui64 uint64
		switch i.(type) {
		case uint:
			ui64 = uint64(i.(uint))
		case uint8:
			ui64 = uint64(i.(uint8))
		case uint16:
			ui64 = uint64(i.(uint16))
		case uint32:
			ui64 = uint64(i.(uint32))
		case uint64:
			ui64 = i.(uint64)
		}
		return strconv.FormatUint(ui64, 10)
	case float32, float64:
		var f64 float64
		switch i.(type) {
		case float32:
			f64 = float64(i.(float32))
		case float64:
			f64 = i.(float64)
		}
		return strconv.FormatFloat(f64, 'G', -1, 64)
	case *float32, *float64:
		var f64 float64
		switch i.(type) {
		case *float32:
			f64 = float64(*i.(*float32))
		case *float64:
			f64 = *i.(*float64)
		}
		return strconv.FormatFloat(f64, 'G', -1, 64)
	case bool:
		return strconv.FormatBool(i.(bool))
	case complex64, complex128:
		var c128 complex128
		switch i.(type) {
		case complex64:
			c128 = complex128(i.(complex64))
		case complex128:
			c128 = i.(complex128)
		}
		return strconv.FormatComplex(c128, 'G', -1, 128)
	case error:
		return i.(error).Error()
	}
	return fmt.Sprintf("%v", i)
}
