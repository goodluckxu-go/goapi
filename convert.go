package goapi

import "strconv"

type Convert string

func (c Convert) String() string {
	return string(c)
}

func (c Convert) Int() int64 {
	val, _ := strconv.ParseInt(string(c), 10, 64)
	return val
}

func (c Convert) Uint() uint64 {
	val, _ := strconv.ParseUint(string(c), 10, 64)
	return val
}

func (c Convert) Float() float64 {
	val, _ := strconv.ParseFloat(string(c), 64)
	return val
}

func (c Convert) Bool() bool {
	val, _ := strconv.ParseBool(string(c))
	return val
}
