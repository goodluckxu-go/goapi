package lang

import (
	"strings"

	"github.com/goodluckxu-go/goapi/v2/utils"
)

type EnUs struct {
}

func (e *EnUs) Required(field string) string {
	return utils.JoinString("The ", field, " is mandatory")
}

func (e *EnUs) Lt(field string, val float64) string {
	return utils.JoinString("The value of ", field, " must be less than ", utils.ToString(val))
}

func (e *EnUs) Lte(field string, val float64) string {
	return utils.JoinString("The value of ", field, " must be less than or equal to ", utils.ToString(val))
}

func (e *EnUs) Gt(field string, val float64) string {
	return utils.JoinString("The value of ", field, " must be greater than ", utils.ToString(val))
}

func (e *EnUs) Gte(field string, val float64) string {
	return utils.JoinString("The value of ", field, " must be greater than or equal to ", utils.ToString(val))
}

func (e *EnUs) MultipleOf(field string, val float64) string {
	return utils.JoinString("The value of ", field, " must be a multiple of ", utils.ToString(val))
}

func (e *EnUs) Max(field string, val uint64) string {
	return utils.JoinString("The maximum length of ", field, " is ", utils.ToString(val))
}

func (e *EnUs) Min(field string, val uint64) string {
	return utils.JoinString("The minimum length of ", field, " is ", utils.ToString(val))
}

func (e *EnUs) Unique(field string) string {
	return utils.JoinString("The value in ", field, " is duplicated")
}

func (e *EnUs) Regexp(field string, val string) string {
	return utils.JoinString("The value of ", field, " does not satisfy the regular expression ", utils.ToString(val))
}

func (e *EnUs) Enum(field string, val []any) string {
	s := ""
	for k, v := range val {
		if k == 0 {
			s += utils.ToString(v)
		} else {
			s += "," + utils.ToString(v)
		}
	}
	return utils.JoinString("The value of ", field, " must be in ", s)
}

func (e *EnUs) JwtTranslate(msg string) string {
	list := strings.Split(msg, ": ")
	for _, v := range jwtErrors {
		if list[0] == v || list[1] == v {
			msg = strings.ToUpper(v[0:1]) + v[1:]
			break
		}
	}
	return msg
}
