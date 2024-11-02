package logger_test

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/iamhectorsosa/anchor/internal/logger"
)

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New()
	log.SetWriter(&buf)

	want := "ANCHOR hello from the logger, key=\"test\" value=\"test_value\"\n"
	log.Info("hello from the logger, key=%q value=%q", "test", "test_value")
	got := stripAnsiCodes(buf.String())

	if got != want {
		t.Errorf("Expected output: %q, but got: %q", want, got)
	}
}

func TestLogger_Error(t *testing.T) {
	log := logger.New()
	want := fmt.Errorf("ERROR err=an error message as a variable")
	got := log.Error("err=%v", "an error message as a variable")

	if got == nil {
		t.Error("Expected an error, but got nil")
	} else if stripAnsiCodes(got.Error()) != want.Error() {
		t.Errorf("Expected error: %q, but got: %q", want, got)
	}
}

func stripAnsiCodes(input string) string {
	re := regexp.MustCompile("\x1b\\[[0-?9;]*[mK]")
	return re.ReplaceAllString(input, "")
}
