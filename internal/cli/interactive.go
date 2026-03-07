package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
)

// isTerminal reports whether the given file descriptor refers to a terminal.
func isTerminal(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// stdoutIsTerminal reports whether stdout is a terminal.
// Extracted for testing.
var stdoutIsTerminal = func() bool {
	return isTerminal(os.Stdout.Fd())
}

// promptDoorSelection displays doors and prompts the user to pick one interactively.
// It reads from reader and writes prompts to writer.
func promptDoorSelection(reader io.Reader, writer io.Writer, doorCount int) (int, error) {
	if _, err := fmt.Fprintf(writer, "\nPick a door (1-%d): ", doorCount); err != nil {
		return 0, fmt.Errorf("write prompt: %w", err)
	}

	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, fmt.Errorf("read input: %w", err)
		}
		return 0, fmt.Errorf("no input received")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return 0, fmt.Errorf("no door selected")
	}

	pick, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid input %q: expected a number between 1 and %d", input, doorCount)
	}

	if pick < 1 || pick > doorCount {
		return 0, fmt.Errorf("invalid door %d: must be between 1 and %d", pick, doorCount)
	}

	return pick, nil
}
