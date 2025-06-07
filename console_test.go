package microconsole

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

const (
	testPrompt       = "Enter input: "
	testConfirmInput = "\n"
)

type failingWriter struct{}

func (w *failingWriter) Write(_ []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write failure")
}

func newTestConsole(input string) *Console {
	return NewWithStreams(strings.NewReader(input), &bytes.Buffer{})
}

func newFailingWriterConsole(input string) *Console {
	return NewWithStreams(strings.NewReader(input), &failingWriter{})
}

func TestConsole_GetInput_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal input", "hello\n", "hello"},
		{"Input with whitespace", "  hello world  \n", "hello world"},
		{"Empty input", "\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := newTestConsole(tt.input)

			result, err := console.GetInput(testPrompt)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected result '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConsole_GetInput_EOF(t *testing.T) {
	console := newTestConsole("")
	_, err := console.GetInput(testPrompt)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestConsole_GetInput_FailedWriter(t *testing.T) {
	console := newFailingWriterConsole("hello\n")
	_, err := console.GetInput(testPrompt)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestConsole_GetConfirm_Success(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		expected   bool
	}{
		{
			name:       "Yes input",
			input:      "y\n",
			defaultYes: false,
			expected:   true,
		},
		{
			name:       "Yes (uppercase) input",
			input:      "Y\n",
			defaultYes: false,
			expected:   true,
		},
		{
			name:       "Yes (full) input",
			input:      "yes\n",
			defaultYes: false,
			expected:   true,
		},
		{
			name:       "Yes (full uppercase) input",
			input:      "YES\n",
			defaultYes: false,
			expected:   true,
		},
		{
			name:       "No input",
			input:      "n\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "No (uppercase) input",
			input:      "N\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "No (full) input",
			input:      "no\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "No (full uppercase) input",
			input:      "NO\n",
			defaultYes: true,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := newTestConsole(tt.input)

			result, err := console.GetConfirm(testPrompt, tt.defaultYes)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected result '%v', got '%v'", tt.expected, result)
			}
		})
	}
}

func TestConsole_GetConfirm_AddsSuffixToPrompt(t *testing.T) {
	console := newTestConsole(testConfirmInput)
	console.GetConfirm(testPrompt, true)
	if !strings.HasSuffix(console.out.(*bytes.Buffer).String(), " [Y/n]: ") {
		t.Error("Expected prompt to end with [Y/n]:")
	}
	console.GetConfirm(testPrompt, false)
	if !strings.HasSuffix(console.out.(*bytes.Buffer).String(), " [y/N]: ") {
		t.Error("Expected prompt to end with [y/N]:")
	}
}

func TestConsole_GetConfirm_DefaultValues(t *testing.T) {
	tests := []struct {
		name       string
		defaultYes bool
		expected   bool
	}{
		{"Empty input with defaultYes=true", true, true},
		{"Empty input with defaultYes=false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := newTestConsole(testConfirmInput)

			result, err := console.GetConfirm(testPrompt, tt.defaultYes)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected result '%v', got '%v'", tt.expected, result)
			}
		})
	}
}

func TestConsole_GetConfirm_Error(t *testing.T) {
	t.Run("Invalid input", func(t *testing.T) {
		console := newTestConsole("invalid\n")

		_, err := console.GetConfirm(testPrompt, false)

		if !errors.Is(err, ErrInvalidConfirmation) {
			t.Errorf("Expected error: %v, got: %v", ErrInvalidConfirmation, err)
		}
	})
}

func TestConsole_GetConfirm_WriterFailure(t *testing.T) {
	t.Run("Failed writer", func(t *testing.T) {
		console := newFailingWriterConsole("y\n")

		_, err := console.GetConfirm(testPrompt, false)

		if err == nil {
			t.Error("Expected error with failing writer, got nil")
		}
	})
}

func TestConsole_GetPassword(t *testing.T) {
	t.Run("Input not os.Stdin", func(t *testing.T) {
		console := newTestConsole("password\n")

		_, err := console.GetPassword(testPrompt)

		if err == nil {
			t.Error("Expected error when using non-os.Stdin for password input, got nil")
		}
	})

	t.Run("Failed writer", func(t *testing.T) {
		console := newFailingWriterConsole("password\n")

		_, err := console.GetPassword(testPrompt)

		if err == nil {
			t.Error("Expected error when writer fails, got nil")
		}
	})
}

func TestConsole_InputWithEOF(t *testing.T) {
	console := newTestConsole("")
	_, err := console.GetInput("Prompt: ")
	if err == nil {
		t.Error("Expected error on EOF, got nil")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("Expected EOF error, got: %v", err)
	}
}

func TestConsole_NilReader(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil reader, but code didn't panic")
		}
	}()

	console := NewWithStreams(nil, &bytes.Buffer{})
	_, _ = console.GetInput("This should panic: ")
}

func TestConsole_NilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil writer, but code didn't panic")
		}
	}()

	console := NewWithStreams(strings.NewReader("input\n"), nil)
	_, _ = console.GetInput("This should panic: ")
}
