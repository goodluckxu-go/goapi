package goapi

// Custom type definition tag interface

type TagRegexp interface {
	Regexp() string
}

type TagEnum interface {
	Enum() []any
}

type TagLt interface {
	Lt() float64
}

type TagLte interface {
	Lte() float64
}

type TagGt interface {
	Gt() float64
}

type TagGte interface {
	Gte() float64
}

type TagMultiple interface {
	Multiple() float64
}

type TagMax interface {
	Max() uint64
}

type TagMin interface {
	Min() uint64
}

type TagUnique interface {
	Unique() bool
}

type TagDesc interface {
	Desc() string
}

type TagDefault interface {
	Default() any
}

type TagExample interface {
	Example() any
}

type TagDeprecated interface {
	Deprecated() bool
}
