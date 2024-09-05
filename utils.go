package goapi

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

func timeFormat(date time.Time, format ...string) string {
	if date.IsZero() {
		return ""
	}
	str := "Y-m-d H:i:s"
	if len(format) > 0 {
		str = format[0]
	}
	year := strconv.Itoa(date.Year())
	month := fmt.Sprintf("%d", date.Month())
	day := strconv.Itoa(date.Day())
	hour := strconv.Itoa(date.Hour())
	minute := strconv.Itoa(date.Minute())
	second := strconv.Itoa(date.Second())
	week := date.Weekday().String()
	weekMap := map[string]string{
		"Monday":    "1",
		"Tuesday":   "2",
		"Wednesday": "3",
		"Thursday":  "4",
		"Friday":    "5",
		"Saturday":  "6",
		"Sunday":    "7",
	}
	str = strings.ReplaceAll(str, "Y", year)
	str = strings.ReplaceAll(str, "m", zeroFill(month, 2))
	str = strings.ReplaceAll(str, "d", zeroFill(day, 2))
	str = strings.ReplaceAll(str, "H", zeroFill(hour, 2))
	str = strings.ReplaceAll(str, "i", zeroFill(minute, 2))
	str = strings.ReplaceAll(str, "s", zeroFill(second, 2))
	str = strings.ReplaceAll(str, "w", weekMap[week])
	str = strings.ReplaceAll(str, "W", week)
	return str
}

func zeroFill(input string, num int) string {
	zeroNum := num - len(input)
	rs := ""
	for i := 0; i < zeroNum; i++ {
		rs += "0"
	}
	return rs + input
}

func spanFill(input string, inputLen, num int) string {
	zeroNum := num - inputLen
	for i := 0; i < zeroNum; i++ {
		input += " "
	}
	return input
}

func IsDefaultLogger(log Logger) (ok bool) {
	var levelLog *levelHandleLogger
	if levelLog, ok = log.(*levelHandleLogger); !ok {
		return
	}
	if levelLog.log == nil {
		return
	}
	_, ok = levelLog.log.(*defaultLogger)
	return
}

func systemBit() int {
	bit := 32 << (^uint(0) >> 63)
	return bit
}
