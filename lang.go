package goapi

type Lang interface {
	Required(field string) string
	Lt(field string, val float64) string
	Lte(field string, val float64) string
	Gt(field string, val float64) string
	Gte(field string, val float64) string
	MultipleOf(field string, val float64) string
	Max(field string, val uint64) string
	Min(field string, val uint64) string
	Unique(field string) string
	Regexp(field string, val string) string
	Enum(field string, val []any) string
}
