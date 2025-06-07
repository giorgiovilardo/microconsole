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
	testPrompt = "Enter input: "
)

type failingWriter struct{}

func (w *failingWriter) Write(_ []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write failure")
}

func newTestConsole(input string) (*Console, *bytes.Buffer) {
	in := strings.NewReader(input)
	out := &bytes.Buffer{}
	return NewWithStreams(in, out), out
}

func newFailingWriterConsole(input string) *Console {
	return NewWithStreams(strings.NewReader(input), &failingWriter{})
}

func TestConsole_GetInput(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
		expectedError bool
		failWriter    bool
	}{
		{
			name:          "Normal input",
			input:         "hello\n",
			expected:      "hello",
			expectedError: false,
			failWriter:    false,
		},
		{
			name:          "Input with whitespace",
			input:         "  hello world  \n",
			expected:      "hello world",
			expectedError: false,
			failWriter:    false,
		},
		{
			name:          "Empty input",
			input:         "\n",
			expected:      "",
			expectedError: false,
			failWriter:    false,
		},
		{
			name:          "EOF",
			input:         "",
			expected:      "",
			expectedError: true,
			failWriter:    false,
		},
		{
			name:          "Failed writer",
			input:         "hello\n",
			expected:      "",
			expectedError: true,
			failWriter:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var console *Console

			if tt.failWriter {
				console = newFailingWriterConsole(tt.input)
			} else {
				console, _ = newTestConsole(tt.input)
			}

			result, err := console.GetInput(testPrompt)

			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}

			if !tt.expectedError && !tt.failWriter {
				if result != tt.expected {
					t.Errorf("Expected result '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestConsole_GetConfirm(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		defaultYes    bool
		expected      bool
		expectedError error
		failWriter    bool
	}{
		{
			name:          "Yes input",
			input:         "y\n",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Yes (uppercase) input",
			input:         "Y\n",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Yes (full) input",
			input:         "yes\n",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Yes (full uppercase) input",
			input:         "YES\n",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "No input",
			input:         "n\n",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "No (uppercase) input",
			input:         "N\n",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "No (full) input",
			input:         "no\n",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "No (full uppercase) input",
			input:         "NO\n",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "Empty input with defaultYes=true",
			input:         "\n",
			defaultYes:    true,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Empty input with defaultYes=false",
			input:         "\n",
			defaultYes:    false,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "Invalid input",
			input:         "invalid\n",
			defaultYes:    false,
			expected:      false,
			expectedError: ErrInvalidConfirmation,
			failWriter:    false,
		},
		{
			name:          "Failed writer",
			input:         "y\n",
			defaultYes:    false,
			expected:      false,
			expectedError: fmt.Errorf("simulated write failure"),
			failWriter:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var console *Console
			var out *bytes.Buffer

			if tt.failWriter {
				console = newFailingWriterConsole(tt.input)
			} else {
				console, out = newTestConsole(tt.input)
			}

			result, err := console.GetConfirm(testPrompt, tt.defaultYes)

			if tt.failWriter {
				if err == nil {
					t.Error("Expected error with failing writer, got nil")
				}
			} else if tt.expectedError == nil {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}

			if !tt.failWriter {
				expectedSuffix := " [Y/n]: "
				if !tt.defaultYes {
					expectedSuffix = " [y/N]: "
				}
				expectedPrompt := testPrompt + expectedSuffix
				if out.String() != expectedPrompt {
					t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, out.String())
				}
			}

			if err == nil && result != tt.expected {
				t.Errorf("Expected result '%v', got '%v'", tt.expected, result)
			}
		})
	}
}

func TestConsole_GetPassword(t *testing.T) {
	t.Run("Input not os.Stdin", func(t *testing.T) {
		console, _ := newTestConsole("password\n")

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
	console, _ := newTestConsole("")
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
