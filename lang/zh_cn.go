package lang

import (
	"fmt"
	"strings"
)

type ZhCn struct {
}

func (z *ZhCn) Required(field string) string {
	return fmt.Sprintf("%v为必填", field)
}

func (z *ZhCn) Lt(field string, val float64) string {
	return fmt.Sprintf("%v的值必须小于%v", field, val)
}

func (z *ZhCn) Lte(field string, val float64) string {
	return fmt.Sprintf("%v的值必须小于等于%v", field, val)
}

func (z *ZhCn) Gt(field string, val float64) string {
	return fmt.Sprintf("%v的值必须大于%v", field, val)
}

func (z *ZhCn) Gte(field string, val float64) string {
	return fmt.Sprintf("%v的值必须大于等于%v", field, val)
}

func (z *ZhCn) MultipleOf(field string, val float64) string {
	return fmt.Sprintf("%v的值必须是%v的倍数", field, val)
}

func (z *ZhCn) Max(field string, val uint64) string {
	return fmt.Sprintf("%v的长度最大值为%v", field, val)
}

func (z *ZhCn) Min(field string, val uint64) string {
	return fmt.Sprintf("%v的长度最小值为%v", field, val)
}

func (z *ZhCn) Unique(field string) string {
	return fmt.Sprintf("%v中的值重复", field)
}

func (z *ZhCn) Regexp(field string, val string) string {
	return fmt.Sprintf("%v的值不满足正则表达式%v", field, val)
}

func (z *ZhCn) Enum(field string, val []any) string {
	s := ""
	for _, v := range val {
		s += fmt.Sprintf(",%v", v)
	}
	if s != "" {
		s = s[1:]
	}
	return fmt.Sprintf("%v的值必须在%v中", field, s)
}

var jwtErrors = []string{
	"key is invalid",
	"key is of invalid type",
	"the requested hash function is unavailable",
	"token is malformed",
	"token is unverifiable",
	"token signature is invalid",
	"token is missing required claim",
	"token has invalid audience",
	"token is expired",
	"token used before issued",
	"token has invalid issuer",
	"token has invalid subject",
	"token is not valid yet",
	"token has invalid id",
	"token has invalid claims",
	"invalid type for claim",
}

func (z *ZhCn) JwtTranslate(msg string) string {
	list := strings.Split(msg, ": ")
	for _, v := range jwtErrors {
		if list[0] == v || list[1] == v {
			msg = v
			break
		}
	}
	zhMap := map[string]string{
		"key is invalid":                             "密钥无效",
		"key is of invalid type":                     "密钥的类型无效",
		"the requested hash function is unavailable": "请求的哈希函数不可用",
		"token is malformed":                         "令牌格式错误",
		"token is unverifiable":                      "令牌不可验证",
		"token signature is invalid":                 "令牌签名无效",
		"token is missing required claim":            "令牌缺少所需的索赔",
		"token has invalid audience":                 "令牌的受众无效",
		"token is expired":                           "令牌过期",
		"token used before issued":                   "发行前使用的令牌",
		"token has invalid issuer":                   "令牌具有无效的颁发者",
		"token has invalid subject":                  "令牌具有无效主题",
		"token is not valid yet":                     "令牌未生效",
		"token has invalid id":                       "令牌的id无效",
		"token has invalid claims":                   "令牌有无效声明",
		"invalid type for claim":                     "无效的索赔类型",
	}
	if zhMap[msg] != "" {
		return zhMap[msg]
	}
	return msg
}
