package goapi

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

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
	fmt.Printf(spanFill(colorDebug("DEBUG"), 5, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Info(format string, a ...any) {
	fmt.Printf(spanFill(colorInfo("INFO"), 4, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Warning(format string, a ...any) {
	fmt.Printf(spanFill(colorWarning("WARNING"), 7, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Error(format string, a ...any) {
	fmt.Printf(spanFill(colorError("ERROR"), 5, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Fatal(format string, a ...any) {
	fmt.Printf(spanFill(colorFatal("FATAL"), 5, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

type levelHandleLogger struct {
	log Logger
}

func (d *levelHandleLogger) Debug(format string, a ...any) {
	if logLevel&LogDebug == 0 {
		return
	}
	d.log.Debug(format, a...)
}

func (d *levelHandleLogger) Info(format string, a ...any) {
	if logLevel&LogInfo == 0 {
		return
	}
	d.log.Info(format, a...)
}

func (d *levelHandleLogger) Warning(format string, a ...any) {
	if logLevel&LogWarning == 0 {
		return
	}
	d.log.Warning(format, a...)
}

func (d *levelHandleLogger) Error(format string, a ...any) {
	if logLevel&LogError == 0 {
		return
	}
	d.log.Error(format, a...)
}

func (d *levelHandleLogger) Fatal(format string, a ...any) {
	if logLevel&LogFail == 0 {
		return
	}
	d.log.Fatal(format, a...)
}

var colorInfo = color.New(color.FgGreen).SprintFunc()
var colorDebug = color.New(color.FgCyan).SprintFunc()
var colorWarning = color.New(color.FgHiYellow).SprintFunc()
var colorError = color.New(color.FgRed).SprintFunc()
var colorFatal = color.New(color.BgRed, color.FgWhite).SprintFunc()

func SetLogLevel(level LogLevel) {
	logLevel = level
}

func SetLogger(log Logger) {
	log = &levelHandleLogger{log: log}
}

func GetLogger() Logger {
	return log
}
