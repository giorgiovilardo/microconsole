# Microconsole

A simple console library for Go.

## Quick Start

### Installation

```bash
go get github.com/giorgiovilardo/microconsole
```

### Basic Usage

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/giorgiovilardo/microconsole"
)

func main() {
    // Get text input
    name, err := microconsole.GetInput("What's your name? ")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Hello, %s!\n", name)
    
    // Get confirmation
    confirmed, err := microconsole.GetConfirm("Continue with the operation?", true)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    if confirmed {
        fmt.Println("Continuing...")
        
        // Get password (masked input)
        password, err := microconsole.GetPassword("Enter your password: ")
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        
        // Never print passwords in a real application!
        fmt.Printf("Password received (length: %d)\n", len(password))
    } else {
        fmt.Println("Operation cancelled.")
    }
}
```

### Advanced Usage with Dependency Injection

For testing or custom I/O streams:

```go
package main

import (
    "fmt"
    "os"
    "strings"
    
    "github.com/giorgiovilardo/microconsole"
)

func main() {
    // Create a custom console for testing
    input := strings.NewReader("test-input\ny\n")
    output := os.Stdout
    console := microconsole.NewWithStreams(input, output)
    
    // Now use the console methods
    text, err := console.GetInput("Enter text: ")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    confirmed, err := console.GetConfirm("Is this correct?", false)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Text: %s, Confirmed: %v\n", text, confirmed)
}
```

## API Reference

### Package Functions

- `GetInput(prompt string) (string, error)` - Prompts for text input
- `GetConfirm(prompt string, defaultYes bool) (bool, error)` - Prompts for yes/no confirmation
- `GetPassword(prompt string) (string, error)` - Prompts for password with masked input

### Console Type

- `New() *Console` - Creates a new console with standard input/output
- `NewWithStreams(in io.Reader, out io.Writer) *Console` - Creates a console with custom I/O
- `(*Console).GetInput(prompt string) (string, error)` - Prompts for text input
- `(*Console).GetConfirm(prompt string, defaultYes bool) (bool, error)` - Prompts for yes/no confirmation
- `(*Console).GetPassword(prompt string) (string, error)` - Prompts for password with masked input

### Errors

- `ErrInvalidConfirmation` - Returned when confirmation input is invalid

## Testing Guidance

The library is designed with testability in mind. Use `NewWithStreams()` to inject mock readers and writers:

```go
func TestMyFunction(t *testing.T) {
    // Setup mock input/output
    input := strings.NewReader("test-input\n")
    output := &bytes.Buffer{}
    console := microconsole.NewWithStreams(input, output)
    
    // Call your function with the console
    result := myFunction(console)
    
    // Assert output was as expected
    if output.String() != "Expected prompt: " {
        t.Errorf("Expected prompt, got: %s", output.String())
    }
    
    // Assert result is correct
    // ...
}
```

## Password Method Limitations

The `GetPassword` method has some limitations:

1. It only works when the input stream is `os.Stdin` - it will return an error otherwise
2. It uses terminal-specific functionality to disable echo
3. It's difficult to test automatically and should be verified manually

For testing code that uses passwords, consider creating an interface that abstracts the console interaction and providing a mock implementation for tests.

## Edge Cases

The library handles several edge cases:

- Empty inputs (returns empty string for `GetInput`, uses default for `GetConfirm`)
- EOF on input (returns appropriate error)
- Invalid confirmation inputs (returns `ErrInvalidConfirmation`)
- Various prompt formats (empty, special characters, very long)

## License

This library is released under the MIT License.