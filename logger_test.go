package goapi

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParseLogFields(t *testing.T) {
	got := ParseLogFields("request_id", "req-001", 42, "status", 200, "lonely")
	want := []LogField{
		{Key: "request_id", Value: "req-001"},
		{Value: 42},
		{Key: "status", Value: 200},
		{Value: "lonely"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseLogFields() = %#v, want %#v", got, want)
	}
}

func TestDefaultLoggerWithFields(t *testing.T) {
	base := &DefaultLogger{Fields: []LogField{{Key: "base", Value: "root"}}}
	logger, ok := base.WithFields("request_id", "req-001", 42).(*DefaultLogger)
	if !ok {
		t.Fatalf("WithFields type = %T, want *DefaultLogger", logger)
	}

	if !reflect.DeepEqual(base.Fields, []LogField{{Key: "base", Value: "root"}}) {
		t.Fatalf("base fields mutated: %#v", base.Fields)
	}

	wantFields := []LogField{
		{Key: "base", Value: "root"},
		{Key: "request_id", Value: "req-001"},
		{Value: 42},
	}
	if !reflect.DeepEqual(logger.Fields, wantFields) {
		t.Fatalf("logger fields = %#v, want %#v", logger.Fields, wantFields)
	}

	output := captureLoggerOutput(t, func() {
		logger.Info("hello %s", "world")
	})
	if !strings.Contains(output, "hello world [base=root,request_id=req-001,!BADKEY=42]") {
		t.Fatalf("logger output missing fields: %q", output)
	}
}

func TestDefaultLoggerWithFieldsNoFieldsReturnsSameLogger(t *testing.T) {
	base := &DefaultLogger{Fields: []LogField{{Key: "base", Value: "root"}}}
	if got := base.WithFields(); got != base {
		t.Fatalf("WithFields without fields should return the same logger")
	}
}

func captureLoggerOutput(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	oldColorful := Colorful
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer func() {
		os.Stdout = oldStdout
		Colorful = oldColorful
	}()

	Colorful = false
	os.Stdout = w
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
