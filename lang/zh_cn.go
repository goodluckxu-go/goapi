package lang

import "fmt"

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
