package lang

import (
	"fmt"
	"strings"
)

type JaJp struct {
}

func (j *JaJp) Required(field string) string {
	return fmt.Sprintf("%vは必須です", field)
}

func (j *JaJp) Lt(field string, val float64) string {
	return fmt.Sprintf("%vの値は%vより小さくなければならない", field, val)
}

func (j *JaJp) Lte(field string, val float64) string {
	return fmt.Sprintf("%vの値は%v以下でなければならない", field, val)
}

func (j *JaJp) Gt(field string, val float64) string {
	return fmt.Sprintf("%v的值必须大于%v", field, val)
}

func (j *JaJp) Gte(field string, val float64) string {
	return fmt.Sprintf("%vの値は%vより大きくなければならない", field, val)
}

func (j *JaJp) MultipleOf(field string, val float64) string {
	return fmt.Sprintf("%vの値は%vの倍数でなければならない", field, val)
}

func (j *JaJp) Max(field string, val uint64) string {
	return fmt.Sprintf("%vの長さの最大値は%v", field, val)
}

func (j *JaJp) Min(field string, val uint64) string {
	return fmt.Sprintf("%vの長さの最小値は%vである", field, val)
}

func (j *JaJp) Unique(field string) string {
	return fmt.Sprintf("%vにおける値の繰り返し", field)
}

func (j *JaJp) Regexp(field string, val string) string {
	return fmt.Sprintf("%vの値が正規表現%vを満たしていない", field, val)
}

func (j *JaJp) Enum(field string, val []any) string {
	s := ""
	for _, v := range val {
		s += fmt.Sprintf(",%v", v)
	}
	if s != "" {
		s = s[1:]
	}
	return fmt.Sprintf("%vの値は%vになければならない", field, s)
}

func (j *JaJp) JwtTranslate(msg string) string {
	list := strings.Split(msg, ": ")
	for _, v := range jwtErrors {
		if list[0] == v || list[1] == v {
			msg = v
			break
		}
	}
	zhMap := map[string]string{
		"key is invalid":                             "鍵が無効です",
		"key is of invalid type":                     "鍵のタイプが無効です",
		"the requested hash function is unavailable": "要求されたハッシュ関数は使用できません",
		"token is malformed":                         "トークンフォーマットエラー",
		"token is unverifiable":                      "トークン認証不可",
		"token signature is invalid":                 "無効なトークン署名",
		"token is missing required claim":            "トークンに必要なクレームがありません",
		"token has invalid audience":                 "無効なトークンの参加者",
		"token is expired":                           "トークンの有効期限",
		"token used before issued":                   "発行前に使用されたトークン",
		"token has invalid issuer":                   "トークンに無効な発行者がいます",
		"token has invalid subject":                  "トークンに無効なトピックがあります",
		"token is not valid yet":                     "トークンが有効ではありません",
		"token has invalid id":                       "トークンのIDが無効です",
		"token has invalid claims":                   "トークンに無効な宣言があります",
		"invalid type for claim":                     "無効なクレームタイプ",
	}
	if zhMap[msg] != "" {
		return zhMap[msg]
	}
	return msg
}

func (j *JaJp) ContentTypeNotSupported(field string) string {
	return fmt.Sprintf("Content-Type値%vはサポートされていません", field)
}
