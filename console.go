package microconsole

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// ErrInvalidConfirmation is returned when a confirmation input is neither yes/y nor no/n
var ErrInvalidConfirmation = errors.New("invalid confirmation input")

// Console provides methods for interacting with the terminal.
// It uses an io.Reader for input and an io.Writer for output.
type Console struct {
	in  io.Reader
	out io.Writer
}

// New creates a new Console instance with standard input and output.
func New() *Console {
	return &Console{os.Stdin, os.Stdout}
}

// NewWithStreams creates a new Console instance with the provided input and output streams.
func NewWithStreams(in io.Reader, out io.Writer) *Console {
	return &Console{in, out}
}

// GetInput writes a prompt to the output and reads a line from the input.
// It trims whitespace from the input before returning.
func (c *Console) GetInput(prompt string) (string, error) {
	_, err := fmt.Fprint(c.out, prompt)
	if err != nil {
		return "", fmt.Errorf("writing prompt: %w", err)
	}

	reader := bufio.NewReader(c.in)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	return strings.TrimSpace(input), nil
}

// GetConfirm prompts the user for a yes/no confirmation and returns a boolean result.
// It accepts "y", "yes", "n", "no" in any case as valid inputs.
// If the input is empty, it returns the defaultYes value.
// If the input is not valid, it returns ErrInvalidConfirmation.
func (c *Console) GetConfirm(prompt string, defaultYes bool) (bool, error) {
	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}

	input, err := c.GetInput(prompt + suffix)
	if err != nil {
		return false, err
	}

	if input == "" {
		return defaultYes, nil
	}

	switch strings.ToLower(input) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, ErrInvalidConfirmation
	}
}

// GetPassword prompts for a password without echoing the input to the terminal.
// Note: This only works when c.in is os.Stdin, as it uses terminal-specific functionality.
func (c *Console) GetPassword(prompt string) (string, error) {
	_, err := fmt.Fprint(c.out, prompt)
	if err != nil {
		return "", fmt.Errorf("writing prompt: %w", err)
	}

	if c.in != os.Stdin {
		return "", fmt.Errorf("password input requires os.Stdin, got different io.Reader")
	}

	password, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}

	fmt.Fprintln(c.out)
	return string(password), nil
}

// Package-level convenience functions using default console
var defaultConsole = New()

// GetInput writes a prompt to the output and reads a line from standard input.
// It trims whitespace from the input before returning.
func GetInput(prompt string) (string, error) {
	return defaultConsole.GetInput(prompt)
}

// GetConfirm prompts the user for a yes/no confirmation and returns a boolean result.
// It accepts "y", "yes", "n", "no" in any case as valid inputs.
// If the input is empty, it returns the defaultYes value.
// If the input is not valid, it returns ErrInvalidConfirmation.
func GetConfirm(prompt string, defaultYes bool) (bool, error) {
	return defaultConsole.GetConfirm(prompt, defaultYes)
}

// GetPassword prompts for a password without echoing the input to the terminal.
// The input is read from standard input.
func GetPassword(prompt string) (string, error) {
	return defaultConsole.GetPassword(prompt)
}
