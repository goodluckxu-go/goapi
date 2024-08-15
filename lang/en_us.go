package lang

import "fmt"

type EnUs struct {
}

func (e *EnUs) Required(field string) string {
	return fmt.Sprintf("The %v is mandatory", field)
}

func (e *EnUs) Lt(field string, val float64) string {
	return fmt.Sprintf("The value of %v must be less than %v", field, val)
}

func (e *EnUs) Lte(field string, val float64) string {
	return fmt.Sprintf("The value of %v must be less than or equal to %v", field, val)
}

func (e *EnUs) Gt(field string, val float64) string {
	return fmt.Sprintf("The value of %v must be greater than %v", field, val)
}

func (e *EnUs) Gte(field string, val float64) string {
	return fmt.Sprintf("The value of %v must be greater than or equal to %v", field, val)
}

func (e *EnUs) MultipleOf(field string, val float64) string {
	return fmt.Sprintf("The value of %v must be a multiple of %v", field, val)
}

func (e *EnUs) Max(field string, val uint64) string {
	return fmt.Sprintf("The maximum length of %v is %v", field, val)
}

func (e *EnUs) Min(field string, val uint64) string {
	return fmt.Sprintf("The minimum length of %v is %v", field, val)
}

func (e *EnUs) Unique(field string) string {
	return fmt.Sprintf("The value in %v is duplicated", field)
}

func (e *EnUs) Regexp(field string, val string) string {
	return fmt.Sprintf("The value of %v does not satisfy the regular expression %v", field, val)
}

func (e *EnUs) Enum(field string, val []any) string {
	s := ""
	for _, v := range val {
		s += fmt.Sprintf(",%v", v)
	}
	if s != "" {
		s = s[1:]
	}
	return fmt.Sprintf("The value of %v must be in %v", field, s)
}
