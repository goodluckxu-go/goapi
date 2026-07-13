package goapi

import (
	"fmt"
	"strings"
	"time"
)

// Logger writes framework logs.
type Logger interface {
	// Debug writes diagnostic details useful while developing or troubleshooting.
	Debug(format string, a ...any)

	// Info writes normal runtime information.
	Info(format string, a ...any)

	// Warn writes a warning that does not stop request handling.
	Warn(format string, a ...any)

	// Error writes an error message for failed operations.
	Error(format string, a ...any)

	// Fatal writes a fatal-level message. Implementations must not terminate the process.
	Fatal(format string, a ...any)

	// WithFields returns a logger enriched with structured fields.
	WithFields(keysAndValues ...any) Logger
}

type LoggerWithContext interface {
	WithContext(ctx *Context) Logger
}

type DefaultLogger struct {
	Fields []LogField
}

func (d *DefaultLogger) Debug(format string, a ...any) {
	d.print(ColorDebug, "DEBUG", format, a...)
}

func (d *DefaultLogger) Info(format string, a ...any) {
	d.print(ColorInfo, "INFO", format, a...)
}

func (d *DefaultLogger) Warn(format string, a ...any) {
	d.print(ColorWarning, "WARNING", format, a...)
}

func (d *DefaultLogger) Error(format string, a ...any) {
	d.print(ColorError, "ERROR", format, a...)
}

func (d *DefaultLogger) Fatal(format string, a ...any) {
	d.print(ColorFatal, "FATAL", format, a...)
}

func (d *DefaultLogger) WithFields(keysAndValues ...any) Logger {
	if len(keysAndValues) == 0 {
		return d
	}
	fields := ParseLogFields(keysAndValues...)
	newFields := make([]LogField, len(d.Fields)+len(fields))
	n := copy(newFields, d.Fields)
	copy(newFields[n:], fields)
	return &DefaultLogger{Fields: newFields}
}

func (d *DefaultLogger) print(colorFun func(a ...any) string, level string, format string, a ...any) {
	format = addStrings(spanFill(colorFun(level), len(level), 10), "[",
		time.Now().Format("2006-01-02 15:04:05"), "] ", format)
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	if len(d.Fields) > 0 {
		var fields []string
		for _, field := range d.Fields {
			key := field.Key
			if key == "" {
				key = "!BADKEY"
			}
			fields = append(fields, key+"="+toString(field.Value))
		}
		format = format + " [" + strings.Join(fields, ",") + "]"
	}
	fmt.Println(format)
}

// ParseLogFields Parse keys and values
// The key must be a string, otherwise, take it as a value
// The value of the key that cannot be matched is the value
func ParseLogFields(keysAndValues ...any) (fields []LogField) {
	fields = make([]LogField, 0, len(keysAndValues))
	for len(keysAndValues) > 0 {
		if len(keysAndValues) > 1 {
			if key, ok := keysAndValues[0].(string); ok {
				fields = append(fields, LogField{key, keysAndValues[1]})
				keysAndValues = keysAndValues[2:]
				continue
			}
		}
		fields = append(fields, LogField{Value: keysAndValues[0]})
		keysAndValues = keysAndValues[1:]
	}
	return
}

var Colorful = true
