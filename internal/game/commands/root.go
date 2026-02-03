package commands

import (
	"strings"

	"fisherevans.com/project/f/internal/game"
	"github.com/spf13/cobra"
)

// Root is a wrapper around cobra.Command to provide easy command registration and execution for game states.
type Root struct {
	command *cobra.Command
}

// New creates a new Root command with the given use and short description.
func New(use, short string) *Root {
	r := &Root{
		command: &cobra.Command{
			Use:   use,
			Short: short,
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
			SilenceErrors: true,
			SilenceUsage:  true,
		},
	}
	r.command.SetOut(game.Console())
	r.command.SetErr(game.Console())
	r.command.Flags().BoolP("help", "h", false, "help for this command")
	r.command.Flags().MarkHidden("help")
	r.command.SetUsageTemplate(`Usage:{{if .HasAvailableSubCommands}}
  [command]{{end}}{{if .HasAvailableLocalFlags}}
  [flags]{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "[command] --help" for more information about a command.{{end}}
`)
	return r
}

// Register adds sub-commands to the root.
func (r *Root) Register(cmds ...*cobra.Command) {
	r.command.AddCommand(cmds...)
}

// HandleInput parses and executes a command string.
func (r *Root) HandleInput(input string) bool {
	args := strings.Split(input, " ")
	if len(args) == 0 || args[0] == "" {
		return false
	}
	r.command.SetArgs(args)
	err := r.command.Execute()
	if err != nil {
		game.Console().Writef("error: %v", err)
	}
	return true
}

// Command returns the underlying cobra.Command.
func (r *Root) Command() *cobra.Command {
	return r.command
}
