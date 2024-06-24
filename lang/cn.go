package lang

import "fmt"

type CN struct {
}

func (c *CN) Required(field string) string {
	return fmt.Sprintf("%v为必填", field)
}

func (c *CN) Lt(field string, val float64) string {
	return fmt.Sprintf("%v的值必须小于%v", field, val)
}

func (c *CN) Lte(field string, val float64) string {
	return fmt.Sprintf("%v的值必须小于等于%v", field, val)
}

func (c *CN) Gt(field string, val float64) string {
	return fmt.Sprintf("%v的值必须大于%v", field, val)
}

func (c *CN) Gte(field string, val float64) string {
	return fmt.Sprintf("%v的值必须大于等于%v", field, val)
}

func (c *CN) MultipleOf(field string, val float64) string {
	return fmt.Sprintf("%v的值必须是%v的倍数", field, val)
}

func (c *CN) Max(field string, val uint64) string {
	return fmt.Sprintf("%v的长度最大值为%v", field, val)
}

func (c *CN) Min(field string, val uint64) string {
	return fmt.Sprintf("%v的长度最小值为%v", field, val)
}

func (c *CN) Unique(field string) string {
	return fmt.Sprintf("%v中的值重复", field)
}

func (c *CN) Regexp(field string, val string) string {
	return fmt.Sprintf("%v的值不满足正则表达式%v", field, val)
}

func (c *CN) Enum(field string, val []any) string {
	return fmt.Sprintf("%v的值必须在%v中", field, val)
}
