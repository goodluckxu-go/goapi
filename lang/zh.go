package lang

var ZhLang = map[string]string{
	Required:   "%v为必填",
	Lt:         "%v的值必须小于%v",
	Lte:        "%v的值必须小于等于%v",
	Gt:         "%v的值必须大于%v",
	Gte:        "%v的值必须大于等于%v",
	MultipleOf: "%v的值必须是%v的倍数",
	Max:        "%v的长度最大值为%v",
	Min:        "%v的长度最小值为%v",
	Unique:     "%v中的值重复",
	Regexp:     "%v的值不满足正则表达式%v",
	Enum:       "%v的值必须在%v中",
}
