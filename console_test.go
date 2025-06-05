package microconsole

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// failingWriter is a writer that always returns an error
type failingWriter struct{}

func (w *failingWriter) Write(_ []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write failure")
}

func TestConsole_GetInput(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		prompt        string
		expected      string
		expectedError bool
		failWriter    bool // This will create a test case where the output writer fails
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
			// Create a Console with a string reader for input and appropriate output
			in := strings.NewReader(tt.input)
			var out io.Writer

			if tt.failWriter {
				// Create a writer that always fails
				out = &failingWriter{}
			} else {
				out = &bytes.Buffer{}
			}

			console := NewWithStreams(in, out)

			// Call GetInput
			result, err := console.GetInput(tt.prompt)

			// Check the error
			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}

			// If no error expected and not using failing writer, check the output and result
			if !tt.expectedError && !tt.failWriter {
				// Check if the prompt was written correctly
				if outBuffer, ok := out.(*bytes.Buffer); ok {
					if outBuffer.String() != tt.prompt {
						t.Errorf("Expected prompt '%s', got '%s'", tt.prompt, outBuffer.String())
					}
				}

				// Check the result
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
			// Create a Console with a string reader for input and appropriate output
			in := strings.NewReader(tt.input)
			var out io.Writer

			if tt.failWriter {
				// Create a writer that always fails
				out = &failingWriter{}
			} else {
				out = &bytes.Buffer{}
			}

			console := NewWithStreams(in, out)

			// Call GetConfirm
			result, err := console.GetConfirm(tt.prompt, tt.defaultYes)

			// Check the error behavior
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

			// If using normal buffer, check the prompt formatting
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

			// Check the result
			if err == nil && result != tt.expected {
				t.Errorf("Expected result '%v', got '%v'", tt.expected, result)
			}
		})
	}
}

func TestConsole_GetPassword(t *testing.T) {
	t.Run("Input not os.Stdin", func(t *testing.T) {
		// Create a Console with a non-os.Stdin reader
		in := strings.NewReader("password\n")
		out := &bytes.Buffer{}
		console := NewWithStreams(in, out)

		// Call GetPassword
		_, err := console.GetPassword("Password: ")

		// Check that we get an error
		if err == nil {
			t.Error("Expected error when using non-os.Stdin for password input, got nil")
		}

		// Check the prompt
		if out.String() != "Password: " {
			t.Errorf("Expected prompt 'Password: ', got '%s'", out.String())
		}
	})

	t.Run("Failed writer", func(t *testing.T) {
		// Create a Console with a failing writer
		in := strings.NewReader("password\n")
		out := &failingWriter{}
		console := NewWithStreams(in, out)

		// Call GetPassword
		_, err := console.GetPassword("Password: ")

		// Check that we get an error
		if err == nil {
			t.Error("Expected error when writer fails, got nil")
		}
	})

	// Note: We can't easily test the actual password reading since it requires terminal interaction
	// This would typically be tested manually or with integration tests
}

func TestGetInput(t *testing.T) {
	// Save and restore defaultConsole
	originalDefault := defaultConsole
	defer func() { defaultConsole = originalDefault }()

	// Mock defaultConsole
	mockIn := strings.NewReader("test input\n")
	mockOut := &bytes.Buffer{}
	defaultConsole = NewWithStreams(mockIn, mockOut)

	// Call the package function
	result, err := GetInput("Enter: ")
	// Verify it worked properly
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
	// Save and restore defaultConsole
	originalDefault := defaultConsole
	defer func() { defaultConsole = originalDefault }()

	// Mock defaultConsole
	mockIn := strings.NewReader("y\n")
	mockOut := &bytes.Buffer{}
	defaultConsole = NewWithStreams(mockIn, mockOut)

	// Call the package function
	result, err := GetConfirm("Confirm?", false)
	// Verify it worked properly
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
	// Save and restore defaultConsole
	originalDefault := defaultConsole
	defer func() { defaultConsole = originalDefault }()

	// Mock defaultConsole with a non-stdin reader
	mockIn := strings.NewReader("password\n")
	mockOut := &bytes.Buffer{}
	defaultConsole = NewWithStreams(mockIn, mockOut)

	// Call the package function
	_, err := GetPassword("Password: ")

	// We expect an error since we're not using os.Stdin
	if err == nil {
		t.Error("Expected error when using non-os.Stdin for password input, got nil")
	}
	if mockOut.String() != "Password: " {
		t.Errorf("Expected prompt 'Password: ', got: '%s'", mockOut.String())
	}
}

// Test edge cases

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

// Examples for documentation

func ExampleGetInput() {
	// This is just an example and won't actually run in tests
	// In a real program, this would prompt the user for input
	// input, err := microconsole.GetInput("What's your name? ")
	// if err != nil {
	//     fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	//     os.Exit(1)
	// }
	// fmt.Printf("Hello, %s!\n", input)
}

func ExampleGetConfirm() {
	// This is just an example and won't actually run in tests
	// In a real program, this would prompt the user for confirmation
	// confirmed, err := microconsole.GetConfirm("Continue?", true)
	// if err != nil {
	//     fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	//     os.Exit(1)
	// }
	// if confirmed {
	//     fmt.Println("Continuing...")
	// } else {
	//     fmt.Println("Operation cancelled.")
	// }
}

func ExampleGetPassword() {
	// This is just an example and won't actually run in tests
	// In a real program, this would prompt the user for a password
	// password, err := microconsole.GetPassword("Enter password: ")
	// if err != nil {
	//     fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	//     os.Exit(1)
	// }
	// fmt.Println("Password received (not showing it for security)")
}

func ExampleConsole_GetInput() {
	// This is just an example and won't actually run in tests
	// In a real program with dependency injection:
	// console := microconsole.NewWithStreams(customInput, customOutput)
	// input, err := console.GetInput("Enter value: ")
	// if err != nil {
	//     // Handle error
	// }
	// // Use input
}

func ExampleConsole_GetConfirm() {
	// This is just an example and won't actually run in tests
	// In a real program with dependency injection:
	// console := microconsole.NewWithStreams(customInput, customOutput)
	// confirmed, err := console.GetConfirm("Proceed?", false)
	// if err != nil {
	//     // Handle error
	// }
	// if confirmed {
	//     // User confirmed
	// } else {
	//     // User declined
	// }
}

func ExampleConsole_GetPassword() {
	// This is just an example and won't actually run in tests
	// Note: This method only works with os.Stdin as input
	// console := microconsole.New()  // Uses os.Stdin and os.Stdout
	// password, err := console.GetPassword("Enter password: ")
	// if err != nil {
	//     // Handle error
	// }
	// // Use password securely
}

// TestConsole_GetInputWithFailingReader tests GetInput with a reader that fails
func TestConsole_GetInputWithFailingReader(t *testing.T) {
	// Create a reader that always fails
	failingReader := &failingReader{}
	console := NewWithStreams(failingReader, &bytes.Buffer{})

	_, err := console.GetInput("Prompt: ")
	if err == nil {
		t.Error("Expected error with failing reader, got nil")
	}
}

// failingReader is a reader that always returns an error
type failingReader struct{}

func (r *failingReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read failure")
}

// TestConsole_SpecificErrorHandling tests specific error handling scenarios
func TestConsole_SpecificErrorHandling(t *testing.T) {
	// Test that errors from GetInput are properly passed through in GetConfirm
	failingReader := &failingReader{}
	console := NewWithStreams(failingReader, &bytes.Buffer{})

	_, err := console.GetConfirm("Confirm?", true)
	if err == nil {
		t.Error("Expected error from GetConfirm with failing reader, got nil")
	}
}

// TestPackageLevelFunctions_ReadPasswordFailure tests the package-level GetPassword function with a mock that fails
func TestPackageLevelFunctions_ReadPasswordFailure(t *testing.T) {
	// Save original values to restore later
	originalDefault := defaultConsole
	originalStdin := mockableStdin

	// Cleanup after test
	defer func() {
		defaultConsole = originalDefault
		mockableStdin = originalStdin
	}()

	// Set up a custom defaultConsole
	defaultConsole = New() // Uses os.Stdin

	// Mock the file descriptor to an invalid value to force failure
	mockableStdin = -1

	// Call GetPassword and expect an error
	_, err := GetPassword("Password: ")
	if err == nil {
		t.Error("Expected error from GetPassword with invalid file descriptor, got nil")
	}
}
