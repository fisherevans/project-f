package commands

import (
	"bytes"
	"strings"
	"testing"

	"fisherevans.com/project/f/internal/game"
	"github.com/spf13/cobra"
)

func TestHelpOutputNotDuplicated(t *testing.T) {
	game.Initialize("test", nil)

	buf := &bytes.Buffer{}
	root := New("", "Root description")
	root.command.SetOut(buf)
	root.command.SetErr(buf)

	root.Register(&cobra.Command{
		Use:   "testcmd",
		Short: "Test command description",
		Run:   func(cmd *cobra.Command, args []string) {},
	})

	// Run help
	root.HandleInput("help")

	output := buf.String()
	t.Logf("Root help output:\n%s", output)

	// Check if "Root description" appears twice
	count := strings.Count(output, "Root description")
	if count != 1 {
		t.Errorf("Expected 'Root description' to appear 1 time, got %d. Output:\n%s", count, output)
	}

	// Reset buffer
	buf.Reset()

	// Run help for subcommand
	root.HandleInput("testcmd --help")
	output = buf.String()
	t.Logf("Subcommand help output:\n%s", output)

	count = strings.Count(output, "Test command description")
	if count != 1 {
		t.Errorf("Expected 'Test command description' to appear 1 time, got %d. Output:\n%s", count, output)
	}

	// Reset buffer
	buf.Reset()

	// Run help [command]
	root.HandleInput("help testcmd")
	output = buf.String()
	t.Logf("help testcmd output:\n%s", output)

	count = strings.Count(output, "Test command description")
	if count != 1 {
		t.Errorf("Expected 'Test command description' to appear 1 time, got %d. Output:\n%s", count, output)
	}
}
