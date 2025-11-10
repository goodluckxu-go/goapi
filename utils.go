package goapi

import (
	"fmt"
	"net"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/goodluckxu-go/goapi/openapi"
)

func inArray[T comparable](val T, list []T) bool {
	for _, v := range list {
		if val == v {
			return true
		}
	}
	return false
}

func inArrayAny(val any, list []any) bool {
	for _, v := range list {
		if val == v {
			return true
		}
	}
	return false
}

func toPtr[T any](val T) *T {
	return &val
}

func spanFill(input string, inputLen, num int) string {
	zeroNum := num - inputLen
	for i := 0; i < zeroNum; i++ {
		input += " "
	}
	return input
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

func isMethod(methods []string) bool {
	for _, method := range methods {
		switch method {
		case http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions, http.MethodHead,
			http.MethodPatch, http.MethodTrace:
		default:
			return false
		}
	}
	return true
}

func isDefaultLogger(log Logger) (ok bool) {
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

func pathJoin(val string, args ...string) string {
	for _, arg := range args {
		val = path.Join(val, arg)
	}
	slash := ""
	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if len(lastArg) > 0 && lastArg[len(lastArg)-1] == '/' {
			slash = "/"
		}
	}
	return val + slash
}

var localIP string

func GetLocalIP() string {
	if localIP == "" {
		conn, err := net.Dial("udp", "114.114.114.114:53")
		if err != nil {
			return "127.0.0.1"
		}
		localIP = conn.LocalAddr().(*net.UDPAddr).IP.String()
	}
	return localIP
}

func removeMorePtr(fType reflect.Type) reflect.Type {
	for fType.Kind() == reflect.Ptr && fType.Elem().Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	return fType
}

func removeAllPtr(fType reflect.Type) reflect.Type {
	for fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	return fType
}

func isArrayType(fType reflect.Type, fn func(sType reflect.Type) bool, deeps ...int) bool {
	deep := -1
	if len(deeps) > 0 {
		deep = deeps[0]
	}
	for {
		if deep == 0 {
			return false
		}
		if deep > 0 {
			deep--
		}
		fType = removeMorePtr(fType)
		if fn(fType) {
			return true
		}
		if fType.Kind() == reflect.Ptr {
			fType = fType.Elem()
			if fn(fType) {
				return true
			}
		}
		if inArray(fType.Kind(), []reflect.Kind{reflect.Array, reflect.Slice}) {
			return isArrayType(fType.Elem(), fn, deep)
		}
		return false
	}
}

func isNormalType(fType reflect.Type) bool {
	switch fType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
		return true
	default:
	}
	return false
}

func isNumberType(fType reflect.Type) bool {
	switch fType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	default:
	}
	return false
}

func toString(v any) string {
	switch val := v.(type) {
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case string:
		return val
	}
	return fmt.Sprintf("%v", v)
}

func toFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func mergeOpenAPITags(tags []*openapi.Tag, args []*openapi.Tag) (list []*openapi.Tag) {
	var tagStrs []string
	for _, tag := range tags {
		if inArray(tag.Name, tagStrs) {
			continue
		}
		tagStrs = append(tagStrs, tag.Name)
		list = append(list, tag)
	}
	for _, tag := range args {
		if inArray(tag.Name, tagStrs) {
			continue
		}
		tagStrs = append(tagStrs, tag.Name)
		list = append(list, tag)
	}
	return
}

// decryptJWT encrypted string based on JWT encryption
func decryptJWT(j *JWT, jwtStr string, bearerJWT HTTPBearerJWT) error {
	pToken, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
		return bearerJWT.DecryptKey()
	})
	if err != nil {
		return err
	}
	if !pToken.Valid {
		return fmt.Errorf("invalid token")
	}
	mapClaims, ok := pToken.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims")
	}
	j.Subject, _ = mapClaims.GetSubject()
	j.Issuer, _ = mapClaims.GetIssuer()
	j.Audience, _ = mapClaims.GetAudience()
	if exp, _ := mapClaims.GetExpirationTime(); exp != nil {
		j.ExpiresAt = exp.Time
	}
	if nbf, _ := mapClaims.GetNotBefore(); nbf != nil {
		j.NotBefore = nbf.Time
	}
	if iat, _ := mapClaims.GetIssuedAt(); iat != nil {
		j.IssuedAt = iat.Time
	}
	if jti, _ := mapClaims["jti"].(string); jti != "" {
		j.ID = jti
	}
	j.Extensions = map[string]any{}
	for k, v := range mapClaims {
		if inArray(k, []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti"}) {
			continue
		}
		j.Extensions[k] = v
	}
	return nil
}
