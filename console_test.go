package microconsole

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"syscall"
	"testing"
)

type failingWriter struct{}

func (w *failingWriter) Write(_ []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write failure")
}

type failingReader struct{}

func (r *failingReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read failure")
}

func TestConsole_GetInput(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		prompt        string
		expected      string
		expectedError bool
		failWriter    bool
	}{
		{
			name:          "Normal input",
			input:         "hello\n",
			prompt:        "Enter input: ",
			expected:      "hello",
			expectedError: false,
			failWriter:    false,
		},
		{
			name:          "Input with whitespace",
			input:         "  hello world  \n",
			prompt:        "Enter input: ",
			expected:      "hello world",
			expectedError: false,
			failWriter:    false,
		},
		{
			name:          "Empty input",
			input:         "\n",
			prompt:        "Enter input: ",
			expected:      "",
			expectedError: false,
			failWriter:    false,
		},
		{
			name:          "EOF",
			input:         "",
			prompt:        "Enter input: ",
			expected:      "",
			expectedError: true,
			failWriter:    false,
		},
		{
			name:          "Failed writer",
			input:         "hello\n",
			prompt:        "Enter input: ",
			expected:      "",
			expectedError: true,
			failWriter:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			var out io.Writer

			if tt.failWriter {
				out = &failingWriter{}
			} else {
				out = &bytes.Buffer{}
			}

			console := NewWithStreams(in, out)
			result, err := console.GetInput(tt.prompt)

			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}

			if !tt.expectedError && !tt.failWriter {
				if outBuffer, ok := out.(*bytes.Buffer); ok {
					if outBuffer.String() != tt.prompt {
						t.Errorf("Expected prompt '%s', got '%s'", tt.prompt, outBuffer.String())
					}
				}

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
		prompt        string
		defaultYes    bool
		expected      bool
		expectedError error
		failWriter    bool
	}{
		{
			name:          "Yes input",
			input:         "y\n",
			prompt:        "Confirm?",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Yes (uppercase) input",
			input:         "Y\n",
			prompt:        "Confirm?",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Yes (full) input",
			input:         "yes\n",
			prompt:        "Confirm?",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Yes (full uppercase) input",
			input:         "YES\n",
			prompt:        "Confirm?",
			defaultYes:    false,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "No input",
			input:         "n\n",
			prompt:        "Confirm?",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "No (uppercase) input",
			input:         "N\n",
			prompt:        "Confirm?",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "No (full) input",
			input:         "no\n",
			prompt:        "Confirm?",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "No (full uppercase) input",
			input:         "NO\n",
			prompt:        "Confirm?",
			defaultYes:    true,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "Empty input with defaultYes=true",
			input:         "\n",
			prompt:        "Confirm?",
			defaultYes:    true,
			expected:      true,
			expectedError: nil,
		},
		{
			name:          "Empty input with defaultYes=false",
			input:         "\n",
			prompt:        "Confirm?",
			defaultYes:    false,
			expected:      false,
			expectedError: nil,
		},
		{
			name:          "Invalid input",
			input:         "invalid\n",
			prompt:        "Confirm?",
			defaultYes:    false,
			expected:      false,
			expectedError: ErrInvalidConfirmation,
			failWriter:    false,
		},
		{
			name:          "Failed writer",
			input:         "y\n",
			prompt:        "Confirm?",
			defaultYes:    false,
			expected:      false,
			expectedError: fmt.Errorf("simulated write failure"),
			failWriter:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			var out io.Writer

			if tt.failWriter {
				out = &failingWriter{}
			} else {
				out = &bytes.Buffer{}
			}

			console := NewWithStreams(in, out)
			result, err := console.GetConfirm(tt.prompt, tt.defaultYes)

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
				out := out.(*bytes.Buffer)
				expectedSuffix := " [Y/n]: "
				if !tt.defaultYes {
					expectedSuffix = " [y/N]: "
				}
				expectedPrompt := tt.prompt + expectedSuffix
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
		in := strings.NewReader("password\n")
		out := &bytes.Buffer{}
		console := NewWithStreams(in, out)

		_, err := console.GetPassword("Password: ")

		if err == nil {
			t.Error("Expected error when using non-os.Stdin for password input, got nil")
		}

		if out.String() != "Password: " {
			t.Errorf("Expected prompt 'Password: ', got '%s'", out.String())
		}
	})

	t.Run("Failed writer", func(t *testing.T) {
		in := strings.NewReader("password\n")
		out := &failingWriter{}
		console := NewWithStreams(in, out)

		_, err := console.GetPassword("Password: ")

		if err == nil {
			t.Error("Expected error when writer fails, got nil")
		}
	})
}

func TestGetInput(t *testing.T) {
	originalDefault := defaultConsole
	defer func() { defaultConsole = originalDefault }()

	mockIn := strings.NewReader("test input\n")
	mockOut := &bytes.Buffer{}
	defaultConsole = NewWithStreams(mockIn, mockOut)

	result, err := GetInput("Enter: ")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "test input" {
		t.Errorf("Expected 'test input', got: '%s'", result)
	}
	if mockOut.String() != "Enter: " {
		t.Errorf("Expected prompt 'Enter: ', got: '%s'", mockOut.String())
	}
}

func TestGetConfirm(t *testing.T) {
	originalDefault := defaultConsole
	defer func() { defaultConsole = originalDefault }()

	mockIn := strings.NewReader("y\n")
	mockOut := &bytes.Buffer{}
	defaultConsole = NewWithStreams(mockIn, mockOut)

	result, err := GetConfirm("Confirm?", false)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !result {
		t.Errorf("Expected true, got: %v", result)
	}
	if mockOut.String() != "Confirm? [y/N]: " {
		t.Errorf("Expected prompt 'Confirm? [y/N]: ', got: '%s'", mockOut.String())
	}
}

func TestGetPassword(t *testing.T) {
	originalDefault := defaultConsole
	defer func() { defaultConsole = originalDefault }()

	mockIn := strings.NewReader("password\n")
	mockOut := &bytes.Buffer{}
	defaultConsole = NewWithStreams(mockIn, mockOut)

	_, err := GetPassword("Password: ")

	if err == nil {
		t.Error("Expected error when using non-os.Stdin for password input, got nil")
	}
	if mockOut.String() != "Password: " {
		t.Errorf("Expected prompt 'Password: ', got: '%s'", mockOut.String())
	}
}

func TestConsole_InputWithEOF(t *testing.T) {
	console := NewWithStreams(strings.NewReader(""), &bytes.Buffer{})
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
	console.GetInput("This should panic: ")
}

func TestConsole_NilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil writer, but code didn't panic")
		}
	}()

	console := NewWithStreams(strings.NewReader("input\n"), nil)
	console.GetInput("This should panic: ")
}

func TestConsole_PromptEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
	}{
		{"Empty prompt", ""},
		{"Special characters", "!@#$%^&*()"},
		{"Very long prompt", strings.Repeat("x", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader("test\n")
			out := &bytes.Buffer{}
			console := NewWithStreams(in, out)

			_, err := console.GetInput(tt.prompt)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if out.String() != tt.prompt {
				t.Errorf("Expected prompt '%s', got '%s'", tt.prompt, out.String())
			}
		})
	}
}

func TestConsole_GetInputWithFailingReader(t *testing.T) {
	failingReader := &failingReader{}
	console := NewWithStreams(failingReader, &bytes.Buffer{})

	_, err := console.GetInput("Prompt: ")
	if err == nil {
		t.Error("Expected error with failing reader, got nil")
	}
}

func TestConsole_SpecificErrorHandling(t *testing.T) {
	failingReader := &failingReader{}
	console := NewWithStreams(failingReader, &bytes.Buffer{})

	_, err := console.GetConfirm("Confirm?", true)
	if err == nil {
		t.Error("Expected error from GetConfirm with failing reader, got nil")
	}
}

func TestPackageLevelFunctions_ReadPasswordFailure(t *testing.T) {
	originalDefault := defaultConsole
	originalStdin := syscall.Stdin

	defer func() {
		defaultConsole = originalDefault
		syscall.Stdin = originalStdin
	}()

	defaultConsole = New()
	syscall.Stdin = -1

	_, err := GetPassword("Password: ")
	if err == nil {
		t.Error("Expected error from GetPassword with invalid file descriptor, got nil")
	}
}