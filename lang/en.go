package lang

var EnLang = map[string]string{
	Required:   "The %v is mandatory",
	Lt:         "The value of %v must be less than %v",
	Lte:        "The value of %v must be less than or equal to %v",
	Gt:         "The value of %v must be greater than %v",
	Gte:        "The value of %v must be greater than or equal to %v",
	MultipleOf: "The value of %v must be a multiple of %v",
	Max:        "The maximum length of %v is %v",
	Min:        "The minimum length of %v is %v",
	Unique:     "The value in %v is duplicated",
	Regexp:     "The value of %v does not satisfy the regular expression %v",
	Enum:       "The value of %v must be in %v",
}
