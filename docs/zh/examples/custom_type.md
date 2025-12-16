## [<<](examples.md) 如何自定义类型
~~~go
// 实现接口
type TextInterface interface {
	UnmarshalText(text []byte) error
	MarshalText() (text []byte, err error)
}
~~~
示例
~~~go
// 重新定义时间类型，使请求和返回都以2006-01-02 15:04:05格式的时间返回
type Time time.Time

func (t *Time) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	tm, err := time.Parse("2006-01-02 15:04:05", string(data))
	if err != nil {
		return err
	}
	*t = Time(tm)
	return nil
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(time.Time(t).Format("2006-01-02 15:04:05")), nil
}
~~~