package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var infoStyle = lipgloss.NewStyle().
	SetString("ANCHOR").
	Bold(true).
	Foreground(lipgloss.Color("80"))

var errorStyle = lipgloss.NewStyle().
	SetString("ERROR").
	Bold(true).
	Foreground(lipgloss.Color("204"))

type Logger struct {
	writer io.Writer
}

func New() *Logger { return &Logger{writer: os.Stdout} }

func (l *Logger) SetWriter(w io.Writer) {
	l.writer = w
}

func (l *Logger) Info(format string, a ...any) {
	fmt.Fprintf(l.writer, "%s %s\n", infoStyle, fmt.Sprintf(format, a...))
}

func (l *Logger) Error(format string, a ...any) error {
	return fmt.Errorf("%s %s", errorStyle, fmt.Sprintf(format, a...))
}
