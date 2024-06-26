package goapi

import "fmt"

type Logger interface {
	Debug(format string, a ...any)
	Info(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
	Fatal(format string, a ...any)
}

type defaultLogger struct {
}

func (d *defaultLogger) Debug(format string, a ...any) {
	fmt.Printf("DEBUG "+format+"\n", a...)
}

func (d *defaultLogger) Info(format string, a ...any) {
	fmt.Printf("INFO "+format+"\n", a...)
}

func (d *defaultLogger) Warning(format string, a ...any) {
	fmt.Printf("WARNING "+format+"\n", a...)
}

func (d *defaultLogger) Error(format string, a ...any) {
	fmt.Printf("ERROR "+format+"\n", a...)
}

func (d *defaultLogger) Fatal(format string, a ...any) {
	fmt.Printf("FATAL "+format+"\n", a...)
}
