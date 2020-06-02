package logging

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LevelString(t *testing.T) {
	tests := map[Level]string{
		DebugLevel:  "debug",
		InfoLevel:   "info",
		WarnLevel:   "warn",
		ErrorLevel:  "error",
		DPanicLevel: "dpanic",
		PanicLevel:  "panic",
		FatalLevel:  "fatal",
		Level(-42):  "Level(-42)",
	}

	for lvl, stringLevel := range tests {
		assert.Equal(t, stringLevel, lvl.String(), "Unexpected lowercase level string.")
		assert.Equal(t, strings.ToUpper(stringLevel), lvl.CapitalString(), "Unexpected all-caps level string.")
	}
}

func Test_LevelText(t *testing.T) {
	tests := []struct {
		text  string
		level Level
	}{
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"", InfoLevel}, // make the zero value useful
		{"warn", WarnLevel},
		{"error", ErrorLevel},
		{"dpanic", DPanicLevel},
		{"panic", PanicLevel},
		{"fatal", FatalLevel},
	}
	for _, tt := range tests {
		if tt.text != "" {
			lvl := tt.level
			marshaled, err := lvl.MarshalText()
			assert.NoError(t, err, "Unexpected error marshaling level %v to text.", &lvl)
			assert.Equal(t, tt.text, string(marshaled), "Marshaling level %v to text yielded unexpected result.", &lvl)
		}

		var unmarshaled Level
		err := unmarshaled.UnmarshalText([]byte(tt.text))
		assert.NoError(t, err, `Unexpected error unmarshaling text %q to level.`, tt.text)
		assert.Equal(t, tt.level, unmarshaled, `Text %q unmarshaled to an unexpected level.`, tt.text)
	}
}

func Test_CapitalLevelsParse(t *testing.T) {
	tests := []struct {
		text  string
		level Level
	}{
		{"DEBUG", DebugLevel},
		{"INFO", InfoLevel},
		{"WARN", WarnLevel},
		{"ERROR", ErrorLevel},
		{"DPANIC", DPanicLevel},
		{"PANIC", PanicLevel},
		{"FATAL", FatalLevel},
	}
	for _, tt := range tests {
		var unmarshaled Level
		err := unmarshaled.UnmarshalText([]byte(tt.text))
		assert.NoError(t, err, `Unexpected error unmarshaling text %q to level.`, tt.text)
		assert.Equal(t, tt.level, unmarshaled, `Text %q unmarshaled to an unexpected level.`, tt.text)
	}
}

func Test_WeirdLevelsParse(t *testing.T) {
	tests := []struct {
		text  string
		level Level
	}{
		// I guess...
		{"Debug", DebugLevel},
		{"Info", InfoLevel},
		{"Warn", WarnLevel},
		{"Error", ErrorLevel},
		{"Dpanic", DPanicLevel},
		{"Panic", PanicLevel},
		{"Fatal", FatalLevel},

		// What even is...
		{"DeBuG", DebugLevel},
		{"InFo", InfoLevel},
		{"WaRn", WarnLevel},
		{"ErRor", ErrorLevel},
		{"DpAnIc", DPanicLevel},
		{"PaNiC", PanicLevel},
		{"FaTaL", FatalLevel},
	}
	for _, tt := range tests {
		var unmarshaled Level
		err := unmarshaled.UnmarshalText([]byte(tt.text))
		assert.NoError(t, err, `Unexpected error unmarshaling text %q to level.`, tt.text)
		assert.Equal(t, tt.level, unmarshaled, `Text %q unmarshaled to an unexpected level.`, tt.text)
	}
}

func Test_LevelNils(t *testing.T) {
	var l *Level

	// The String() method will not handle nil level properly.
	assert.Panics(t, func() {
		assert.Equal(t, "Level(nil)", l.String(), "Unexpected result stringifying nil *Level.")
	}, "Level(nil).String() should panic")

	assert.Panics(t, func() {
		l.MarshalText()
	}, "Expected to panic when marshalling a nil level.")

	err := l.UnmarshalText([]byte("debug"))
	assert.Equal(t, errUnmarshalNilLevel, err, "Expected to error unmarshalling into a nil Level.")
}

func Test_LevelUnmarshalUnknownText(t *testing.T) {
	var l Level
	err := l.UnmarshalText([]byte("foo"))
	assert.Contains(t, err.Error(), "unrecognized level", "Expected unmarshaling arbitrary text to fail.")
}

func Test_LevelAsFlagValue(t *testing.T) {
	var (
		buf bytes.Buffer
		lvl Level
	)
	fs := flag.NewFlagSet("levelTest", flag.ContinueOnError)
	fs.SetOutput(&buf)
	fs.Var(&lvl, "level", "log level")

	for _, expected := range []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, FatalLevel} {
		assert.NoError(t, fs.Parse([]string{"-level", expected.String()}))
		assert.Equal(t, expected, lvl, "Unexpected level after parsing flag.")
		assert.Empty(t, buf.String(), "Unexpected error output parsing level flag.")
		buf.Reset()
	}

	assert.Error(t, fs.Parse([]string{"-level", "nope"}))
	assert.Equal(
		t,
		`invalid value "nope" for flag -level: unrecognized level: "nope"`,
		strings.Split(buf.String(), "\n")[0], // second line is help message
		"Unexpected error output from invalid flag input.",
	)
}
