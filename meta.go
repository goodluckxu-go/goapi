package goapi

// Custom type definition meta interface

type MetaRegexp interface {
	Regexp() string
}

type MetaEnum interface {
	Enum() []any
}

type MetaLt interface {
	Lt() float64
}

type MetaLte interface {
	Lte() float64
}

type MetaGt interface {
	Gt() float64
}

type MetaGte interface {
	Gte() float64
}

type MetaMultiple interface {
	Multiple() float64
}

type MetaMax interface {
	Max() uint64
}

type MetaMin interface {
	Min() uint64
}

type MetaUnique interface {
	Unique() bool
}

type MetaDesc interface {
	Desc() string
}

type MetaDefault interface {
	Default() any
}

type MetaExample interface {
	Example() any
}

type MetaDeprecated interface {
	Deprecated() bool
}

type MetaName interface {
	Name() string
}

type MetaValidate interface {
	Validate() error
}
